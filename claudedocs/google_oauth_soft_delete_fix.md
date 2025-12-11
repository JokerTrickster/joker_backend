# Google OAuth Soft-Delete Account Restoration Fix

**Date:** 2025-12-11
**Status:** Completed
**Issue:** Soft-deleted users unable to login via Google OAuth
**Solution:** Automatic account restoration on re-login

---

## Problem Statement

### Symptom
Users who had previously deleted their accounts (soft-delete) were unable to login again via Google OAuth. The system returned "Failed to find or create user" error.

### Root Cause
The `FindOrCreateUserByEmail()` method in `googleSigninAuthRepository.go` was not using `.Unscoped()` when querying the database. GORM's soft-delete feature automatically filters out records where `deleted_at IS NOT NULL`, so deleted users were invisible to the query.

When a deleted user tried to login:
1. System queried for user by email (without `.Unscoped()`)
2. Query returned no results (deleted user filtered out)
3. System attempted to create new user with same email
4. Database constraint violation (email must be unique)
5. Error: "Failed to find or create user"

### Impact
- Deleted users could not re-activate their accounts via Google login
- Poor user experience
- Customer support burden

---

## Solution Implementation

### Changes Made

**File:** `/Users/luxrobo/project/joker_backend/services/authService/features/googleSignin/repository/googleSigninAuthRepository.go`

**Before:**
```go
func (r *GoogleSigninAuthRepository) FindOrCreateUserByEmail(email, name, provider string) (*models.User, error) {
    var user models.User

    // Query without .Unscoped() - deleted users filtered out
    if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            newUser := models.User{
                Email:    email,
                Name:     name,
                Provider: provider,
            }
            if err := r.DB.Create(&newUser).Error; err != nil {
                return nil, fmt.Errorf("failed to create user: %v", err)
            }
            return &newUser, nil
        }
        return nil, err
    }

    return &user, nil
}
```

**After:**
```go
func (r *GoogleSigninAuthRepository) FindOrCreateUserByEmail(email, name, provider string) (*models.User, error) {
    var user models.User

    // Use .Unscoped() to include soft-deleted records
    if err := r.DB.Unscoped().Where("email = ?", email).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            newUser := models.User{
                Email:    email,
                Name:     name,
                Provider: provider,
            }
            if err := r.DB.Create(&newUser).Error; err != nil {
                return nil, fmt.Errorf("failed to create user: %v", err)
            }
            return &newUser, nil
        }
        return nil, err
    }

    // Restore soft-deleted user by clearing DeletedAt
    if user.DeletedAt.Valid {
        user.DeletedAt = gorm.DeletedAt{}
        if err := r.DB.Unscoped().Save(&user).Error; err != nil {
            return nil, fmt.Errorf("failed to restore user: %v", err)
        }
    }

    return &user, nil
}
```

### Key Changes
1. Added `.Unscoped()` to include soft-deleted records in query
2. Check if user is soft-deleted (`user.DeletedAt.Valid`)
3. Restore user by clearing `DeletedAt` field
4. Use `.Unscoped().Save()` to update soft-deleted record

---

## Behavior

### New User Flow
1. User logs in via Google OAuth (first time)
2. System queries database with `.Unscoped()`
3. No user found → Create new user
4. Return user with JWT tokens

### Existing User Flow
1. User logs in via Google OAuth (return visit)
2. System queries database with `.Unscoped()`
3. User found with `deleted_at = NULL`
4. Return user with JWT tokens

### Deleted User Restoration Flow
1. Previously deleted user logs in via Google OAuth
2. System queries database with `.Unscoped()`
3. User found with `deleted_at != NULL`
4. System clears `deleted_at` field → Account restored
5. Return user with JWT tokens

**User Data Preservation:**
- User ID remains the same
- `created_at` timestamp preserved
- All user metadata preserved
- No duplicate accounts created

---

## Testing

### Test Suite Created
**File:** `/Users/luxrobo/project/joker_backend/services/authService/features/googleSignin/repository/googleSigninAuthRepository_test.go`

**Test Scenarios:**
1. `TestFindOrCreateUserByEmail_CreateNewUser` - New user creation
2. `TestFindOrCreateUserByEmail_FindExistingUser` - Existing user lookup
3. `TestFindOrCreateUserByEmail_RestoreSoftDeletedUser` - Deleted user restoration

### Test Results
```bash
=== RUN   TestFindOrCreateUserByEmail_CreateNewUser
--- PASS: TestFindOrCreateUserByEmail_CreateNewUser

=== RUN   TestFindOrCreateUserByEmail_FindExistingUser
--- PASS: TestFindOrCreateUserByEmail_FindExistingUser

=== RUN   TestFindOrCreateUserByEmail_RestoreSoftDeletedUser
--- PASS: TestFindOrCreateUserByEmail_RestoreSoftDeletedUser

PASS
ok      github.com/JokerTrickster/joker_backend/services/authService/features/googleSignin/repository
```

All tests passing, confirming:
- New user creation works correctly
- Existing user lookup works correctly
- Soft-deleted user restoration works correctly
- No duplicate accounts created
- User data preserved during restoration

---

## Security Considerations

### Access Control
- Only users with valid Google ID tokens can trigger restoration
- Email must exactly match the soft-deleted account
- Google OAuth token validation ensures authenticity

### Data Integrity
- Original user ID preserved (no foreign key violations)
- Created timestamp preserved (audit trail maintained)
- Provider field remains 'google'
- No data loss during restoration

### Privacy
- Restoration is automatic and transparent to user
- No notification sent (same as normal login)
- Existing user data remains private

---

## Documentation Updates

### Files Updated
1. **GOOGLE_OAUTH_SETUP.md** - Added section on account restoration behavior
2. **GOOGLE_LOGIN_TEST.md** - Added test scenario for soft-deleted account restoration
3. **claudedocs/google_oauth_soft_delete_fix.md** - This document (technical summary)

### Key Documentation Points
- Soft-delete restoration is automatic and transparent
- No user action required beyond normal Google login
- All user data preserved during restoration
- No duplicate accounts created
- Security controls in place

---

## Future Considerations

### Potential Enhancements
1. **Notification System** - Optional email notification when account is restored
2. **Audit Logging** - Log restoration events for compliance
3. **Restoration Limits** - Prevent abuse with rate limiting
4. **User Choice** - Allow users to permanently delete vs soft-delete

### Monitoring
- Track restoration rate in analytics
- Monitor for restoration abuse patterns
- Alert on unusual restoration spikes

---

## References

### Related Files
- `/services/authService/features/googleSignin/repository/googleSigninAuthRepository.go`
- `/services/authService/features/googleSignin/repository/googleSigninAuthRepository_test.go`
- `/shared/models/user.go`
- `/GOOGLE_OAUTH_SETUP.md`
- `/GOOGLE_LOGIN_TEST.md`

### GORM Documentation
- [Soft Delete](https://gorm.io/docs/delete.html#Soft-Delete)
- [Unscoped Queries](https://gorm.io/docs/delete.html#Find-soft-deleted-records)

---

**Implementation Date:** 2025-12-11
**Tested By:** Automated test suite
**Deployed To:** Development environment
**Production Deployment:** Pending

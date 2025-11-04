/**
 * API 엔드포인트 템플릿
 * 이 템플릿을 복사하여 새로운 API 엔드포인트를 만들 때 사용하세요.
 */

// Express.js 예시
const express = require('express');
const router = express.Router();

// GET 요청 예시
router.get('/:id', async (req, res) => {
  try {
    // 1. 요청 파라미터 검증
    const { id } = req.params;
    if (!id) {
      return res.status(400).json({
        success: false,
        error: {
          code: 'INVALID_PARAMS',
          message: 'ID parameter is required'
        }
      });
    }

    // 2. 비즈니스 로직 실행
    // const data = await service.getById(id);

    // 3. 성공 응답
    return res.status(200).json({
      success: true,
      data: {}, // 실제 데이터
      message: 'Resource retrieved successfully'
    });

  } catch (error) {
    // 4. 에러 핸들링
    console.error('Error:', error);
    return res.status(500).json({
      success: false,
      error: {
        code: 'INTERNAL_ERROR',
        message: 'An internal error occurred'
      }
    });
  }
});

// POST 요청 예시
router.post('/', async (req, res) => {
  try {
    // 1. 요청 본문 검증
    const { name, email } = req.body;
    if (!name || !email) {
      return res.status(400).json({
        success: false,
        error: {
          code: 'INVALID_DATA',
          message: 'Name and email are required'
        }
      });
    }

    // 2. 비즈니스 로직 실행
    // const newResource = await service.create({ name, email });

    // 3. 성공 응답 (201 Created)
    return res.status(201).json({
      success: true,
      data: {}, // 새로 생성된 리소스
      message: 'Resource created successfully'
    });

  } catch (error) {
    // 4. 에러 핸들링
    console.error('Error:', error);
    return res.status(500).json({
      success: false,
      error: {
        code: 'INTERNAL_ERROR',
        message: 'An internal error occurred'
      }
    });
  }
});

module.exports = router;


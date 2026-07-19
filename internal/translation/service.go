package translation

import "context"

type Service struct {
	translator Translator
	cache      *Cache
}

func NewService(translator Translator) *Service {
	return &Service{
		translator: translator,
		cache:      NewCache(),
	}
}

func (s *Service) Translate(ctx context.Context, req Request) (string, error) {
	if cached, ok := s.cache.Get(req.Direction, req.Message, req.Context); ok {
		return cached, nil
	}

	result, err := s.translator.Translate(ctx, req)
	if err != nil {
		return "", err
	}

	s.cache.Set(req.Direction, req.Message, req.Context, result)
	return result, nil
}

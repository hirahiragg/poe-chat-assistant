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

func (s *Service) SetTranslator(t Translator) {
	s.translator = t
	s.cache = NewCache()
}

func (s *Service) Translate(ctx context.Context, req Request) (string, error) {
	if cached, ok := s.cache.Get(req); ok {
		return cached, nil
	}

	result, err := s.translator.Translate(ctx, req)
	if err != nil {
		return "", err
	}

	s.cache.Set(req, result)
	return result, nil
}

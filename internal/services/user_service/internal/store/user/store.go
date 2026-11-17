package user

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/pkg/postgres"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	el "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/elasticsearch"
	mn "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/minio"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg/meta"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store/user/pg/user"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/minio/minio-go/v7"
)

type Store struct {
	metaRepository   store.MetaRepository
	userRepository   store.UserRepository
	searchRepository store.SearchRepository
	photosRepository store.PhotosRepository
}

func New(
	cfg Config,
	logger logger.Logger,
	pgDB *postgres.Postgres,
	mnDB *minio.Client,
	mnPresigner *minio.Client,
	elDB *elasticsearch.Client,
) (store.Store, error) {
	logger = logger.WithFields("layer", "store")

	metaRepository := meta.New(pgDB, logger)
	userRepository := user.New(pgDB, logger)

	photosRepository := mn.New(mnDB, mnPresigner, logger, cfg.Minio.Bucket, cfg.Minio.URLExpiry)

	searchRepository := el.New(elDB, logger, cfg.Elasticsearch.DoctorIndex)

	return &Store{
		metaRepository:   metaRepository,
		userRepository:   userRepository,
		photosRepository: photosRepository,
		searchRepository: searchRepository,
	}, nil
}

func (s *Store) User() store.UserRepository {
	return s.userRepository
}

func (s *Store) Meta() store.MetaRepository {
	return s.metaRepository
}

func (s *Store) Search() store.SearchRepository {
	return s.searchRepository
}

func (s *Store) Photos() store.PhotosRepository {
	return s.photosRepository
}

func (s *Store) Close() error {
	return s.searchRepository.Close()
}

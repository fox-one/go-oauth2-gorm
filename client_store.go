package oauth2gorm

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/jinzhu/gorm"
)

type ClientStoreItem struct {
	ID        string
	Secret    string `gorm:"type:varchar(512)"`
	Domain    string `gorm:"type:varchar(512)"`
	Data      string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	table string `gorm:"-"`
}

func (p ClientStoreItem) TableName() string {
	if p.table != "" {
		return p.table
	}

	return "oauth2_clients"
}

func NewClientStoreWithDB(config *Config, db *gorm.DB) *ClientStore {
	store := &ClientStore{
		db:     db,
		stdout: os.Stderr,
	}

	csi := &ClientStoreItem{table: config.TableName}
	store.tableName = csi.TableName()

	if !db.HasTable(store.tableName) {
		if err := db.CreateTable(csi).Error; err != nil {
			panic(err)
		}
	}

	return store
}

type ClientStore struct {
	tableName string
	db        *gorm.DB
	stdout    io.Writer
}

func (s *ClientStore) toClientInfo(data []byte) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := json.Unmarshal(data, &cm)
	return &cm, err
}

func (s *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var item ClientStoreItem
	err := s.db.Table(s.tableName).Limit(1).Find(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return s.toClientInfo([]byte(item.Data))
}

func (s *ClientStore) Create(ctx context.Context, info oauth2.ClientInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	item := &ClientStoreItem{
		ID:     info.GetID(),
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		Data:   string(data),
	}

	return s.db.Table(s.tableName).Create(item).Error
}

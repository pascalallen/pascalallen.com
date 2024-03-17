package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

type Session interface {
	First(dest interface{}, conds ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	Create(value interface{}) error
	Delete(value interface{}, conds ...interface{}) error
	Save(value interface{}) error
	Preload(query string, args ...interface{}) Session
	Where(query interface{}, args ...interface{}) Session
	Replace(model interface{}, association string, values ...interface{}) error
	AutoMigrate(dests ...interface{})
}

type GormSession struct {
	session *gorm.DB
}

func NewGormSession() Session {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to initialize database session: %s", err)
	}

	return GormSession{session: db}
}

func (s GormSession) AutoMigrate(dests ...interface{}) {
	if err := s.session.AutoMigrate(dests...); err != nil {
		log.Fatalf("failed to auto migrate database: %s", err)
	}
}

func (s GormSession) First(dest interface{}, conds ...interface{}) error {
	if err := s.session.First(dest, conds...).Error; err != nil {
		return err
	}

	return nil
}

func (s GormSession) Find(dest interface{}, conds ...interface{}) error {
	if err := s.session.Find(dest, conds...).Error; err != nil {
		return err
	}

	return nil
}

func (s GormSession) Create(value interface{}) error {
	if err := s.session.Create(value).Error; err != nil {
		return err
	}

	return nil
}

func (s GormSession) Delete(value interface{}, conds ...interface{}) error {
	if err := s.session.Delete(value, conds...).Error; err != nil {
		return err
	}

	return nil
}

func (s GormSession) Save(value interface{}) error {
	if err := s.session.Save(value).Error; err != nil {
		return err
	}

	return nil
}

func (s GormSession) Preload(query string, args ...interface{}) Session {
	s.session = s.session.Preload(query, args...)

	return s
}

func (s GormSession) Where(query interface{}, args ...interface{}) Session {
	s.session = s.session.Where(query, args...)

	return s
}

func (s GormSession) Replace(model interface{}, association string, values ...interface{}) error {
	if err := s.session.Model(model).Association(association).Replace(values...); err != nil {
		return err
	}

	return nil
}

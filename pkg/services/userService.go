package services

import (
	"emelyanenkoig/reminder/pkg/models"
	"emelyanenkoig/reminder/pkg/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.repo.CreateUser(user)
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.repo.GetUserById(id)
}

func (s *UserService) UpdateUser(user *models.User) error {
	return s.repo.UpdateUser(user.ID, user)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.repo.DeleteUser(id)
}

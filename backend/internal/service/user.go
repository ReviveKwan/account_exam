package service

import (
	"accounting-exam/internal/config"
	"accounting-exam/internal/model"
	"accounting-exam/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repository.UserRepo
	cfg      *config.Config
}

func NewUserService(userRepo *repository.UserRepo, cfg *config.Config) *UserService {
	return &UserService{userRepo: userRepo, cfg: cfg}
}

// LoginByWeChat 微信登录
func (s *UserService) LoginByWeChat(code string) (*model.User, string, error) {
	openid := s.code2OpenID(code)
	if openid == "" {
		return nil, "", fmt.Errorf("微信登录凭证无效")
	}

	user, err := s.userRepo.FindByOpenID(openid)
	if err != nil {
		if !repository.IsNotFound(err) {
			return nil, "", fmt.Errorf("查询用户失败: %v", err)
		}
		// 新用户自动注册
		user = &model.User{
			ID:       "u_" + uuid.New().String(),
			OpenID:   openid,
			Nickname: "",
			IsGuest:  0,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, "", fmt.Errorf("用户创建失败")
		}
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// LoginAsGuest 游客登录（免登录试刷）
func (s *UserService) LoginAsGuest() (*model.User, string, error) {
	openid := "guest_" + uuid.New().String()[:12]
	user := &model.User{
		ID:       "u_" + uuid.New().String(),
		OpenID:   openid,
		Nickname: "试刷学员",
		IsGuest:  1,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("游客创建失败: %v", err)
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

func (s *UserService) generateToken(user *model.User) (string, error) {
	return generateTokenInternal(user.ID, user.OpenID, s.cfg.JWT.Secret)
}

func (s *UserService) code2OpenID(code string) string {
	if s.cfg.WeChat.AppID != "" && s.cfg.WeChat.Secret != "" {
		// TODO: 生产环境调用微信 jscode2session 接口换取真实 openid
		return ""
	}
	if code == "" {
		return ""
	}
	return "dev_openid_" + code
}

package services

import (
	"database/sql"
	"errors"
	"fmt"
	"melogo/internal/model"
	"melogo/internal/utils"
	"time"
)

// UserService 用户服务
type UserService struct {
	db *sql.DB
}

// NewUserService 创建用户服务实例
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// Register 用户注册
func (us *UserService) Register(username, email, password string) (*model.User, error) {
	// 检查用户名是否已存在
	exists, err := us.UsernameExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if email != "" {
		exists, err = us.EmailExists(email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("邮箱已被使用")
		}
	}

	// 检查是否是第一个注册的用户，如果是则设为管理员
	userCount, err := us.GetUserCount()
	if err != nil {
		return nil, fmt.Errorf("检查用户数量失败: %v", err)
	}

	isAdmin := 0
	if userCount == 0 {
		// 第一个注册的用户设为管理员
		isAdmin = 1
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 插入用户数据
	query := `
		INSERT INTO users (username, password_hash, email, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := us.db.Exec(query, username, hashedPassword, email, isAdmin, time.Now(), time.Now())
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	// 获取插入的用户ID
	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 返回创建的用户信息
	user := &model.User{
		ID:        int(userID),
		Username:  username,
		Email:     email,
		IsAdmin:   isAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return user, nil
}

// Login 用户登录
func (us *UserService) Login(username, password string) (*model.User, string, error) {
	// 根据用户名查询用户
	user, err := us.GetUserByUsername(username)
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 验证密码
	err = utils.VerifyPassword(user.Password, password)
	if err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %v", err)
	}

	// 清空密码字段，避免返回给前端
	user.Password = ""

	return user, token, nil
}

// GetUserByID 根据ID获取用户信息
func (us *UserService) GetUserByID(id int) (*model.User, error) {
	query := `
		SELECT id, username, email, avatar, is_admin, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	var user model.User
	var email, avatar sql.NullString
	err := us.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&email,
		&avatar,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 处理NULL值
	if email.Valid {
		user.Email = email.String
	}
	if avatar.Valid {
		user.Avatar = avatar.String
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户信息（包含密码hash）
func (us *UserService) GetUserByUsername(username string) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, is_admin, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	var user model.User
	var avatar sql.NullString
	err := us.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&avatar,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 处理NULL值
	if avatar.Valid {
		user.Avatar = avatar.String
	}

	return &user, nil
}

// UsernameExists 检查用户名是否存在
func (us *UserService) UsernameExists(username string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = ?"
	err := us.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// EmailExists 检查邮箱是否存在
func (us *UserService) EmailExists(email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE email = ?"
	err := us.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserCount 获取用户总数
func (us *UserService) GetUserCount() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM users"
	err := us.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateUser 更新用户信息
func (us *UserService) UpdateUser(id int, email, avatar string) error {
	query := `
		UPDATE users
		SET email = ?, avatar = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := us.db.Exec(query, email, avatar, time.Now(), id)
	if err != nil {
		return fmt.Errorf("更新用户信息失败: %v", err)
	}
	return nil
}

// UpdateUserAvatarData 更新用户头像数据(BLOB)
func (us *UserService) UpdateUserAvatarData(id int, avatarData []byte) error {
	query := `
		UPDATE users
		SET avatar_blob = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := us.db.Exec(query, avatarData, time.Now(), id)
	if err != nil {
		return fmt.Errorf("更新用户头像失败: %v", err)
	}
	return nil
}

// GetUserAvatarData 获取用户头像数据(BLOB)
func (us *UserService) GetUserAvatarData(id int) ([]byte, error) {
	query := `SELECT avatar_blob FROM users WHERE id = ?`
	var avatarData []byte
	// Use sql.NullString or byte slice scanning
	// Note: If blobl is NULL, Scan might fail if we don't use *[]byte or valid check.
	// However, []byte handles NULL as nil/empty usually in standard driver, let's verify.
	// Actually safer to use *[]byte or sql.RawBytes
	err := us.db.QueryRow(query, id).Scan(&avatarData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return avatarData, nil
}

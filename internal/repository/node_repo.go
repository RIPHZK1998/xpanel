package repository

import (
	"errors"
	"time"

	"xpanel/internal/models"

	"gorm.io/gorm"
)

// Node errors
var (
	ErrNodeNotFound = errors.New("node not found")
)

// NodeRepository handles node database operations.
type NodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository creates a new node repository instance.
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Create inserts a new node into the database.
func (r *NodeRepository) Create(node *models.Node) error {
	return r.db.Create(node).Error
}

// GetByID retrieves a node by its ID.
func (r *NodeRepository) GetByID(id uint) (*models.Node, error) {
	var node models.Node
	result := r.db.First(&node, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNodeNotFound
		}
		return nil, result.Error
	}
	return &node, nil
}

// GetAll retrieves all nodes.
func (r *NodeRepository) GetAll() ([]models.Node, error) {
	var nodes []models.Node
	result := r.db.Find(&nodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return nodes, nil
}

// GetOnlineNodes retrieves all online nodes based on recent heartbeat.
// A node is considered online if it has sent a heartbeat within the last 2 minutes.
func (r *NodeRepository) GetOnlineNodes() ([]models.Node, error) {
	var nodes []models.Node
	// Consider a node online if heartbeat was within last 2 minutes
	twoMinutesAgo := time.Now().Add(-2 * time.Minute)
	result := r.db.Where("last_check_at > ?", twoMinutesAgo).Find(&nodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return nodes, nil
}

// GetAvailableNodes retrieves all nodes that can accept new users.
func (r *NodeRepository) GetAvailableNodes() ([]models.Node, error) {
	var nodes []models.Node
	result := r.db.Where("status = ? AND (max_users = 0 OR current_users < max_users)",
		models.NodeStatusOnline).Find(&nodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return nodes, nil
}

// Update updates node fields.
func (r *NodeRepository) Update(node *models.Node) error {
	return r.db.Save(node).Error
}

// UpdateStatus updates the node status.
func (r *NodeRepository) UpdateStatus(id uint, status models.NodeStatus) error {
	return r.db.Model(&models.Node{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateLastCheck updates the last health check timestamp.
func (r *NodeRepository) UpdateLastCheck(id uint) error {
	now := time.Now()
	return r.db.Model(&models.Node{}).Where("id = ?", id).Update("last_check_at", &now).Error
}

// UpdateHeartbeat updates only the heartbeat-related fields (status, current_users, last_check_at).
// If realityPublicKey is non-empty, it also updates the node's Reality public key.
// This is much faster than updating all node fields.
func (r *NodeRepository) UpdateHeartbeat(id uint, status models.NodeStatus, currentUsers int, lastCheckAt *time.Time, realityPublicKey string) error {
	updates := map[string]interface{}{
		"status":        status,
		"current_users": currentUsers,
		"last_check_at": lastCheckAt,
	}

	// Only update Reality public key if provided
	if realityPublicKey != "" {
		updates["reality_public_key"] = realityPublicKey
	}

	return r.db.Model(&models.Node{}).Where("id = ?", id).Updates(updates).Error
}

// IncrementUserCount increments the current user count for a node.
func (r *NodeRepository) IncrementUserCount(id uint) error {
	return r.db.Model(&models.Node{}).Where("id = ?", id).
		UpdateColumn("current_users", gorm.Expr("current_users + 1")).Error
}

// DecrementUserCount decrements the current user count for a node.
func (r *NodeRepository) DecrementUserCount(id uint) error {
	return r.db.Model(&models.Node{}).Where("id = ?", id).
		UpdateColumn("current_users", gorm.Expr("GREATEST(current_users - 1, 0)")).Error
}

// Delete soft-deletes a node.
func (r *NodeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Node{}, id).Error
}

// GetByProtocol retrieves nodes by protocol type.
func (r *NodeRepository) GetByProtocol(protocol models.ProtocolType) ([]models.Node, error) {
	var nodes []models.Node
	result := r.db.Where("protocol = ? AND status = ?", protocol, models.NodeStatusOnline).Find(&nodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return nodes, nil
}

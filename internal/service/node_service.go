package service

import (
	"time"
	"xpanel/internal/models"
	"xpanel/internal/repository"
)

// NodeService handles node-related business logic.
type NodeService struct {
	nodeRepo     *repository.NodeRepository
	activityRepo *repository.ActivityRepository
}

// NewNodeService creates a new node service.
func NewNodeService(
	nodeRepo *repository.NodeRepository,
	activityRepo *repository.ActivityRepository,
) *NodeService {
	return &NodeService{
		nodeRepo:     nodeRepo,
		activityRepo: activityRepo,
	}
}

// GetAllNodes retrieves all VPN nodes.
func (s *NodeService) GetAllNodes() ([]models.Node, error) {
	nodes, err := s.nodeRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Populate online devices count
	for i := range nodes {
		count, err := s.activityRepo.CountOnlineUsersByNode(nodes[i].ID, 2*time.Minute)
		if err == nil {
			nodes[i].OnlineDevices = int(count)
			// Also update CurrentUsers to reflect reality if needed, or just use OnlineDevices for display
			nodes[i].CurrentUsers = int(count)
		}
	}

	return nodes, nil
}

// GetOnlineNodes retrieves all online nodes.
func (s *NodeService) GetOnlineNodes() ([]models.Node, error) {
	return s.nodeRepo.GetOnlineNodes()
}

// GetAvailableNodes retrieves nodes that can accept new users.
func (s *NodeService) GetAvailableNodes() ([]models.Node, error) {
	return s.nodeRepo.GetAvailableNodes()
}

// GetNode retrieves a specific node by ID.
func (s *NodeService) GetNode(id uint) (*models.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Populate online devices count
	count, err := s.activityRepo.CountOnlineUsersByNode(node.ID, 2*time.Minute)
	if err == nil {
		node.OnlineDevices = int(count)
		node.CurrentUsers = int(count)
	}

	return node, nil
}

// CreateNode creates a new VPN node.
func (s *NodeService) CreateNode(node *models.Node) error {
	return s.nodeRepo.Create(node)
}

// UpdateNode updates an existing node.
func (s *NodeService) UpdateNode(node *models.Node) error {
	return s.nodeRepo.Update(node)
}

// UpdateNodeStatus updates a node's status.
func (s *NodeService) UpdateNodeStatus(id uint, status models.NodeStatus) error {
	return s.nodeRepo.UpdateStatus(id, status)
}

// DeleteNode removes a node.
func (s *NodeService) DeleteNode(id uint) error {
	return s.nodeRepo.Delete(id)
}

// GetNodesByProtocol retrieves nodes by protocol type.
func (s *NodeService) GetNodesByProtocol(protocol models.ProtocolType) ([]models.Node, error) {
	return s.nodeRepo.GetByProtocol(protocol)
}

// IncrementNodeUsers increments the user count for a node.
func (s *NodeService) IncrementNodeUsers(id uint) error {
	return s.nodeRepo.IncrementUserCount(id)
}

// DecrementNodeUsers decrements the user count for a node.
func (s *NodeService) DecrementNodeUsers(id uint) error {
	return s.nodeRepo.DecrementUserCount(id)
}

package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type Node struct {
	gorm.Model
	Name               string `json:"name" gorm:"not null"`
	DaemonTaskNum      uint   `json:"daemonTaskNum"`
	Disabled           bool   `json:"disabled"` // 通信失败时Disabled会被设置为true
	CrontabTaskNum     uint   `json:"crontabTaskNum"`
	GroupID            uint   `json:"groupID" gorm:"not null;unique_index:uni_group_addr" `
	CrontabJobAuditNum uint   `json:"crontabJobAuditNum"`
	DaemonJobAuditNum  uint   `json:"daemonJobAuditNum"`
	CrontabJobFailNum  uint   `json:"crontabJobFailNum"`
	Addr               string `json:"addr" gorm:"not null;unique_index:uni_group_addr"`
	Group              Group  `json:"group"`
}

func (n *Node) VerifyUserGroup(userID, groupID uint, addr string) bool {
	var user User

	if groupID == SuperGroup.ID {
		return true
	}

	if DB().Take(&user, "id=? and group_id=?", userID, groupID).Error != nil {
		return false
	}

	return n.Exists(groupID, addr)
}

func (n *Node) Delete(groupID uint, addr string) error {
	var ret *gorm.DB
	DB().Take(n, "group_id=? and addr=?", groupID, addr)
	if groupID == SuperGroup.ID {
		// 超级管理员分组采用软删除
		ret = DB().Delete(n, "group_id=? and addr=?", groupID, addr)
	} else {
		ret = DB().Unscoped().Delete(n, "group_id=? and addr=?", groupID, addr)
	}

	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected == 0 {
		return errors.New("Delete failed")
	}
	return nil
}

func (n *Node) Rename(groupID uint, addr string) error {
	return DB().Model(n).Where("group_id=? and addr=?", groupID, addr).Updates(n).Error
}

// GroupNode 为节点分组，复制groupID=1分组中node至目标分组
func (n *Node) GroupNode(addr string, targetGroupID uint, targetNodeName, targetGroupName string) error {

	// 新建分组
	if targetGroupID == 0 {
		group := &Group{
			Name: targetGroupName,
		}
		if err := DB().Save(group).Error; err != nil {
			return err
		}
		targetGroupID = group.ID
	}

	err := DB().Preload("Group").Where("group_id=? and addr=?", SuperGroup.ID, addr).Take(n).Error
	if err != nil {
		return err
	}

	if targetNodeName == "" {
		targetNodeName = n.Name
	}

	return DB().Save(&Node{
		Addr:    addr,
		GroupID: targetGroupID,
		Name:    targetNodeName,
	}).Error
}

func (n *Node) Exists(groupID uint, addr string) bool {
	if DB().Take(n, "group_id=? and addr=?", groupID, addr).Error != nil {
		return false
	}
	return true
}

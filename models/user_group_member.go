package models

import (
	"jditms/pkg/ctx"
	"jditms/pkg/poster"
)

type UserGroupMember struct {
	GroupId int64
	UserId  int64
}

func (UserGroupMember) TableName() string {
	return "user_group_member"
}

func MyGroupIds(ctx *ctx.Context, userId int64) ([]int64, error) {
	var ids []int64
	err := DB(ctx).Model(&UserGroupMember{}).Where("user_id=?", userId).Pluck("group_id", &ids).Error
	return ids, err
}

// my business group ids
func MyBusiGroupIds(ctx *ctx.Context, userId int64) ([]int64, error) {
	groupIds, err := MyGroupIds(ctx, userId)
	if err != nil {
		return []int64{}, err
	}

	return BusiGroupIds(ctx, groupIds)
}

func MemberIds(ctx *ctx.Context, groupId int64) ([]int64, error) {
	var ids []int64
	err := DB(ctx).Model(&UserGroupMember{}).Where("group_id=?", groupId).Pluck("user_id", &ids).Error
	return ids, err
}

func UserGroupMemberCount(ctx *ctx.Context, where string, args ...interface{}) (int64, error) {
	return Count(DB(ctx).Model(&UserGroupMember{}).Where(where, args...))
}

func UserGroupMemberAdd(ctx *ctx.Context, groupId, userId int64) error {
	num, err := UserGroupMemberCount(ctx, "user_id=? and group_id=?", userId, groupId)
	if err != nil {
		return err
	}

	if num > 0 {
		// already exists
		return nil
	}

	obj := UserGroupMember{
		GroupId: groupId,
		UserId:  userId,
	}

	return Insert(ctx, obj)
}

func UserGroupMemberDel(ctx *ctx.Context, groupId int64, userIds []int64) error {
	if len(userIds) == 0 {
		return nil
	}

	return DB(ctx).Where("group_id = ? and user_id in ?", groupId, userIds).Delete(&UserGroupMember{}).Error
}

func UserGroupMemberGetAll(ctx *ctx.Context) ([]*UserGroupMember, error) {
	if !ctx.IsCenter {
		lst, err := poster.GetByUrls[[]*UserGroupMember](ctx, "/v1/jditms/user-group-members")
		return lst, err
	}

	var lst []*UserGroupMember
	err := DB(ctx).Find(&lst).Error
	return lst, err
}

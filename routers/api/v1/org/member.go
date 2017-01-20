// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package org

import (
	"fmt"

	api "code.gitea.io/sdk/gitea"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/routers/api/v1/user"
)

// listMembers list an organization's members
func listMembers(ctx *context.APIContext, publicOnly bool) {
	var members []*models.User
	if publicOnly {
		orgUsers, err := models.GetOrgUsersByOrgID(ctx.Org.Organization.ID)
		if err != nil {
			ctx.Error(500, "GetOrgUsersByOrgID", err)
			return
		}

		memberIDs := make([]int64, 0, len(orgUsers))
		for _, orgUser := range orgUsers {
			if orgUser.IsPublic {
				memberIDs = append(memberIDs, orgUser.UID)
			}
		}

		if members, err = models.GetUsersByIDs(memberIDs); err != nil {
			ctx.Error(500, "GetUsersByIDs", err)
			return
		}
	} else {
		if err := ctx.Org.Organization.GetMembers(); err != nil {
			ctx.Error(500, "GetMembers", err)
			return
		}
		members = ctx.Org.Organization.Members
	}

	apiMembers := make([]*api.User, len(members))
	for i, member := range members {
		apiMembers[i] = member.APIFormat()
	}
	ctx.JSON(200, apiMembers)
}

// ListMembers list an organization's members
func ListMembers(ctx *context.APIContext) {
	listMembers(ctx, !ctx.Org.Organization.IsOrgMember(ctx.User.ID))
}

// ListPublicMembers list an organization's public members
func ListPublicMembers(ctx *context.APIContext) {
	listMembers(ctx, true)
}

// IsMember check if a user is a member of an organization
func IsMember(ctx *context.APIContext) {
	org := ctx.Org.Organization
	requester := ctx.User
	userToCheck := user.GetUserByParams(ctx)
	if org.IsOrgMember(requester.ID) {
		if org.IsOrgMember(userToCheck.ID) {
			ctx.Status(204)
		} else {
			ctx.Status(404)
		}
	} else if requester.ID == userToCheck.ID {
		ctx.Status(404)
	} else {
		redirectURL := fmt.Sprintf("%sapi/v1/orgs/%s/public_members/%s",
			setting.AppURL, org.Name, userToCheck.Name)
		ctx.Redirect(redirectURL, 302)
	}
}

// IsPublicMember check if a user is a public member of an organization
func IsPublicMember(ctx *context.APIContext) {
	userToCheck := user.GetUserByParams(ctx)
	if userToCheck.IsPublicMember(ctx.Org.Organization.ID) {
		ctx.Status(204)
	} else {
		ctx.Status(404)
	}
}

// PublicizeMember make a member's membership public
func PublicizeMember(ctx *context.APIContext) {
	userToPublicize := user.GetUserByParams(ctx)
	if userToPublicize.ID != ctx.User.ID {
		ctx.Error(403, "", "Cannot publicize another member")
		return
	} else if !ctx.Org.Organization.IsOrgMember(userToPublicize.ID) {
		ctx.Error(403, "", "Must be a member of the organization")
		return
	}
	err := models.ChangeOrgUserStatus(ctx.Org.Organization.ID, userToPublicize.ID, true)
	if err != nil {
		ctx.Error(500, "ChangeOrgUserStatus", err)
		return
	}
	ctx.Status(204)
}

// ConcealMember make a member's membership not public
func ConcealMember(ctx *context.APIContext) {
	userToConceal := user.GetUserByParams(ctx)
	if userToConceal.ID != ctx.User.ID {
		ctx.Error(403, "", "Cannot conceal another member")
		return
	} else if !ctx.Org.Organization.IsOrgMember(userToConceal.ID) {
		ctx.Error(403, "", "Must be a member of the organization")
		return
	}
	err := models.ChangeOrgUserStatus(ctx.Org.Organization.ID, userToConceal.ID, false)
	if err != nil {
		ctx.Error(500, "ChangeOrgUserStatus", err)
		return
	}
	ctx.Status(204)
}

// DeleteMember remove a member from an organization
func DeleteMember(ctx *context.APIContext) {
	org := ctx.Org.Organization
	if !org.IsOwnedBy(ctx.User.ID) {
		ctx.Error(403, "", "You must be an owner of the organization.")
		return
	}
	if err := org.RemoveMember(user.GetUserByParams(ctx).ID); err != nil {
		ctx.Error(500, "RemoveMember", err)
	}
	ctx.Status(204)
}

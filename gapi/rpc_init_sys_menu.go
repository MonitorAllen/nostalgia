package gapi

import (
	"context"
	"errors"
	"github.com/MonitorAllen/nostalgia/internal/cache"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const adminInitMenuKey = "admin:menus:"

func (server *Server) InitSysMenu(ctx context.Context, _ *pb.InitSysMenuRequest) (*pb.InitSysMenuResponse, error) {
	accessPayload, _, err := server.authorizeAdmin(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	adminMenuKey := cache.GetAdminMenuKey(accessPayload.RoleID)

	// 从 redis 中获取
	var treeInitSysMenuList []*pb.InitSysMenu
	ok, err := server.cache.Get(ctx, adminMenuKey, &treeInitSysMenuList)
	if err != nil && !errors.Is(redis.Nil, err) {
		return nil, status.Errorf(codes.Internal, "get admin init menu: %v", err)
	}

	if ok {
		return &pb.InitSysMenuResponse{
			InitSysMenu: treeInitSysMenuList,
		}, nil
	}

	initSysMenus, err := server.store.ListInitSysMenus(ctx, accessPayload.RoleID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can not get init system menus: %v", err)
	}

	initSysMenuList := convertInitSysMenu(initSysMenus)

	treeInitSysMenuList = buildMenuTree(initSysMenuList)

	err = server.cache.Set(ctx, adminMenuKey, treeInitSysMenuList, time.Hour*12)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can not cache init system menu: %v", err)
	}

	return &pb.InitSysMenuResponse{
		InitSysMenu: treeInitSysMenuList,
	}, nil
}

// 构建树状结构并填充 Children 字段
func buildMenuTree(menus []*pb.InitSysMenu) []*pb.InitSysMenu {
	menuMap := make(map[int64]*pb.InitSysMenu)
	var rootMenus []*pb.InitSysMenu

	// Step 1: 将所有菜单放入 map 中
	for _, menu := range menus {
		menuMap[menu.Id] = menu // 使用原始结构体，避免拷贝
	}

	// Step 2: 构建树形结构
	for _, menu := range menus {
		if *menu.ParentId == 0 { // ParentId 为 0 代表根节点
			rootMenus = append(rootMenus, menu)
		} else {
			parent, found := menuMap[*menu.ParentId]
			if found {
				// 将当前菜单加入父菜单的 Children
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return rootMenus
}

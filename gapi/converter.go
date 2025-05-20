package gapi

import (
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertAdmin(admin db.Admin) *pb.Admin {
	return &pb.Admin{
		Username:  admin.Username,
		IsActive:  admin.IsActive,
		CreatedAt: timestamppb.New(admin.CreatedAt),
	}
}

func convertInitSysMenu(sysMenus []db.ListInitSysMenusRow) []*pb.InitSysMenu {
	sysMenuList := make([]*pb.InitSysMenu, 0)
	for _, menu := range sysMenus {
		pbMenu := &pb.InitSysMenu{
			Id:       menu.ID,
			Name:     menu.Name,
			Path:     &menu.Path, // 即使为空字符串，也会被赋值
			Icon:     menu.Icon,
			ParentId: &menu.ParentID.Int64, // 根节点 ParentId 可能为 0
		}
		sysMenuList = append(sysMenuList, pbMenu)
	}
	return sysMenuList
}

func convertArticle(article db.ListAllArticlesRow) *pb.Article {
	return &pb.Article{
		Id:        article.ID.String(),
		Title:     article.Title,
		Summary:   &article.Summary,
		IsPublish: &article.IsPublish,
		Views:     &article.Views,
		Likes:     &article.Likes,
		CreatedAt: timestamppb.New(article.CreatedAt),
		UpdatedAt: timestamppb.New(article.UpdatedAt),
		DeletedAt: timestamppb.New(article.DeletedAt),
		Owner:     article.Owner.String(),
	}
}

func convertArticleList(articles []db.ListAllArticlesRow) []*pb.Article {
	articlesList := make([]*pb.Article, 0)
	if len(articles) == 0 {
		return articlesList
	}
	for _, article := range articles {
		articlesList = append(articlesList, convertArticle(article))
	}
	return articlesList
}

func convertOnlyArticle(article db.Article) *pb.Article {
	return &pb.Article{
		Id:        article.ID.String(),
		Title:     article.Title,
		Summary:   &article.Summary,
		IsPublish: &article.IsPublish,
		Views:     &article.Views,
		Likes:     &article.Likes,
		CreatedAt: timestamppb.New(article.CreatedAt),
		UpdatedAt: timestamppb.New(article.UpdatedAt),
		DeletedAt: timestamppb.New(article.DeletedAt),
		Owner:     article.Owner.String(),
	}
}

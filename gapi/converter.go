package gapi

import (
	db "github.com/MonitorAllen/nostalgia/db/sqlc"
	"github.com/MonitorAllen/nostalgia/pb"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertArticle(article db.ListAllArticlesRow) *pb.Article {
	return &pb.Article{
		Id:                  article.ID.String(),
		Title:               article.Title,
		Summary:             &article.Summary,
		IsPublish:           &article.IsPublish,
		Views:               &article.Views,
		Likes:               &article.Likes,
		Slug:                article.Slug.String,
		CheckOutdated:       &article.CheckOutdated,
		ReadTime:            article.ReadTime,
		LastUpdated:         timestamppb.New(article.LastUpdated),
		CreatedByAutomation: &article.CreatedByAutomation,
		AutomationStatus:    article.AutomationStatus,
		CreatedAt:           timestamppb.New(article.CreatedAt),
		UpdatedAt:           timestamppb.New(article.UpdatedAt),
		DeletedAt:           timestamppb.New(article.DeletedAt),
		Owner:               article.Owner.String(),
		CategoryName:        article.CategoryName.String,
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

func convertOnlyArticle(article db.Article, needContent bool) *pb.Article {
	pbArticle := &pb.Article{
		Id:         article.ID.String(),
		Title:      article.Title,
		Summary:    &article.Summary,
		IsPublish:  &article.IsPublish,
		Views:      &article.Views,
		Likes:      &article.Likes,
		CreatedAt:  timestamppb.New(article.CreatedAt),
		UpdatedAt:  timestamppb.New(article.UpdatedAt),
		DeletedAt:  timestamppb.New(article.DeletedAt),
		Owner:      article.Owner.String(),
		CategoryId: article.CategoryID,
	}
	if needContent {
		pbArticle.Content = &article.Content
	}
	return pbArticle
}

func convertArticleWithCategory(article db.GetArticleRow, needContent bool) *pb.Article {
	pbArticle := &pb.Article{
		Id:            article.ID.String(),
		Title:         article.Title,
		Summary:       &article.Summary,
		Views:         &article.Views,
		Likes:         &article.Likes,
		IsPublish:     &article.IsPublish,
		Slug:          article.Slug.String,
		CheckOutdated: &article.CheckOutdated,
		ReadTime:      article.ReadTime,
		LastUpdated:   timestamppb.New(article.LastUpdated),
		CreatedAt:     timestamppb.New(article.CreatedAt),
		UpdatedAt:     timestamppb.New(article.UpdatedAt),
		Owner:         article.Owner.String(),
		CategoryId:    article.CategoryID,
		CategoryName:  article.CategoryName.String,
	}
	if needContent {
		pbArticle.Content = &article.Content
	}
	return pbArticle
}

func convertCategory(category db.Category) *pb.Category {
	return &pb.Category{
		Id:        category.ID,
		Name:      category.Name,
		IsSystem:  category.IsSystem,
		CreatedAt: timestamppb.New(category.CreatedAt),
		UpdatedAt: timestamppb.New(category.UpdatedAt),
	}
}

func convertCategories(categories []db.Category) []*pb.Category {
	categoriesList := make([]*pb.Category, 0)
	for _, category := range categories {
		pbCate := convertCategory(category)
		categoriesList = append(categoriesList, pbCate)
	}

	return categoriesList
}

func convertCategoryCountArticleRow(category db.ListCategoriesCountArticlesRow) *pb.Category {
	return &pb.Category{
		Id:           category.ID,
		Name:         category.Name,
		IsSystem:     category.IsSystem,
		ArticleCount: &category.ArticleCount,
		CreatedAt:    timestamppb.New(category.CreatedAt),
		UpdatedAt:    timestamppb.New(category.UpdatedAt),
	}
}
func convertCategoriesCountArticleRow(categories []db.ListCategoriesCountArticlesRow) []*pb.Category {
	categoriesList := make([]*pb.Category, 0)
	for _, category := range categories {
		pbCate := convertCategoryCountArticleRow(category)
		categoriesList = append(categoriesList, pbCate)
	}

	return categoriesList
}

func optionalTimestamp(value pgtype.Timestamptz) *timestamppb.Timestamp {
	if !value.Valid {
		return nil
	}
	return timestamppb.New(value.Time)
}

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Id:              user.ID.String(),
		Username:        user.Username,
		FullName:        user.FullName,
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		Role:            user.Role,
		CreatedAt:       timestamppb.New(user.CreatedAt),
		UpdatedAt:       timestamppb.New(user.UpdatedAt),
		DisabledAt:      optionalTimestamp(user.DisabledAt),
		DisabledReason:  user.DisabledReason,
	}
}

func convertAdminUserRow(user db.ListAdminUsersRow) *pb.User {
	return &pb.User{
		Id:              user.ID.String(),
		Username:        user.Username,
		FullName:        user.FullName,
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		Role:            user.Role,
		CreatedAt:       timestamppb.New(user.CreatedAt),
		UpdatedAt:       timestamppb.New(user.UpdatedAt),
		DisabledAt:      optionalTimestamp(user.DisabledAt),
		DisabledReason:  user.DisabledReason,
	}
}

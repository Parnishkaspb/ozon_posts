package comments

import (
	"context"
	"errors"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrPostIDRequired    = errors.New("postID is required")
	ErrMax2000Symbols    = errors.New("text max 2000 symbols")
	ErrCommentIDRequired = errors.New("commentID is required")
	ErrCantWriteComment  = errors.New("can't write a comment to this post")
)

type CommentRepo interface {
	CreateComment(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error)
	AnswerComment(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) error
}

type PostRepo interface {
	WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error)
}

type CommentService struct {
	commentRepo CommentRepo
	postRepo    PostRepo
}

func New(comment CommentRepo, post PostRepo) *CommentService {
	return &CommentService{
		commentRepo: comment,
		postRepo:    post,
	}
}

func (s *CommentService) normalizeAndValidateText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("text is required")
	}
	if len([]rune(text)) > 2000 {
		return "", ErrMax2000Symbols
	}
	return text, nil
}

func (s *CommentService) requireUUID(id uuid.UUID, err error) error {
	if id == uuid.Nil {
		return err
	}
	return nil
}

func (s *CommentService) CommentCreate(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error) {
	text, err := s.normalizeAndValidateText(text)
	if err != nil {
		return &models.Comment{}, err
	}
	if err := s.requireUUID(authorID, errors.New("authorID is required")); err != nil {
		return &models.Comment{}, err
	}
	if err := s.requireUUID(postID, ErrPostIDRequired); err != nil {
		return &models.Comment{}, err
	}

	ok, err := s.postRepo.WithoutComment(ctx, postID)
	if err != nil {
		return &models.Comment{}, err
	}

	if !ok {
		return &models.Comment{}, ErrCantWriteComment
	}

	return s.commentRepo.CreateComment(ctx, text, authorID, postID)
}

func (s *CommentService) CommentAnswer(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) error {
	text, err := s.normalizeAndValidateText(text)
	if err != nil {
		return err
	}
	if err := s.requireUUID(authorID, errors.New("authorID is required")); err != nil {
		return err
	}
	if err := s.requireUUID(postID, ErrPostIDRequired); err != nil {
		return err
	}
	if err := s.requireUUID(commentID, ErrCommentIDRequired); err != nil {
		return err
	}

	return s.commentRepo.AnswerComment(ctx, text, authorID, postID, commentID)
}

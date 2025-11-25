package repository

import (
	"database/sql"
	"digit/internal/model"
)

type BookRepository interface {
	List() ([]model.Book, error)
	Create(req model.CreateBookRequest) (uint64, error)
	Get(id uint64) (*model.Book, error)
	Update(id uint64, req model.UpdateBookRequest) error
	Delete(id uint64) error
}

type MySQLBookRepository struct {
	db *sql.DB
}

func NewMySQLBookRepository(db *sql.DB) *MySQLBookRepository {
	return &MySQLBookRepository{db: db}
}

func (r *MySQLBookRepository) List() ([]model.Book, error) {
	rows, err := r.db.Query("SELECT id, title, author, published_year, isbn, created_at, updated_at FROM books")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var books []model.Book
	for rows.Next() {
		var b model.Book
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.PublishedYear, &b.ISBN, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (r *MySQLBookRepository) Create(req model.CreateBookRequest) (uint64, error) {
	stmt, err := r.db.Prepare("INSERT INTO books (title, author, published_year, isbn) VALUES (?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(req.Title, req.Author, req.PublishedYear, req.ISBN)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *MySQLBookRepository) Get(id uint64) (*model.Book, error) {
	row := r.db.QueryRow("SELECT id, title, author, published_year, isbn, created_at, updated_at FROM books WHERE id = ?", id)
	var b model.Book
	err := row.Scan(&b.ID, &b.Title, &b.Author, &b.PublishedYear, &b.ISBN, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *MySQLBookRepository) Update(id uint64, req model.UpdateBookRequest) error {
	stmt, err := r.db.Prepare("UPDATE books SET title = ?, author = ?, published_year = ?, isbn = ? WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(req.Title, req.Author, req.PublishedYear, req.ISBN, id)
	return err
}

func (r *MySQLBookRepository) Delete(id uint64) error {
	stmt, err := r.db.Prepare("DELETE FROM books WHERE id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id)
	return err
}

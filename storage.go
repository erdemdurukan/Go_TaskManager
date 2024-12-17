package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	createUser(*User) error
	getAllUser() ([]*User, error)
	getUserById(int) (*User, error)
	FindByEmail(string) (*User, error)
	createItem(*Item) error
	getItemsofUser(int) ([]*Item, error)
	changeItemStatus(int, string, int) error
	deleteItem(int, int) error
	getItemsByCreator(int) ([]*Item, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=password123 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to the database!")

	return &PostgresStore{
		db: db,
	}, nil
}

func (p *PostgresStore) Close() {
	p.db.Close()
}

func (s *PostgresStore) Init() (error, error) {
	return s.createUserTable(), s.createItemTable()
}
func (s *PostgresStore) createItemTable() error {
	// Enum tipi oluştur
	enumQuery := `
		DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'item_status') THEN
				CREATE TYPE item_status AS ENUM ('open', 'in progress', 'closed');
			END IF;
		END$$;
	`
	_, err := s.db.Exec(enumQuery)
	if err != nil {
		return fmt.Errorf("could not create enum type: %w", err)
	}

	// Tablo oluştur
	query := `
		CREATE TABLE IF NOT EXISTS items (
			id SERIAL PRIMARY KEY,
			creatorId int NOT NULL,
			itemOwnerId int NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			status item_status NOT NULL,
			message VARCHAR(100) NOT NULL,
			isDeleted BOOLEAN NOT NULL DEFAULT FALSE
		);
	`
	_, err = s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("could not create items table: %w", err)
	}

	return nil
}

// working like repository in java
func (s *PostgresStore) createUserTable() error {
	query := `
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(50) NOT NULL,
            password VARCHAR(100) NOT NULL,
            email VARCHAR(100) UNIQUE NOT NULL
        );
    `
	_, err := s.db.Exec(query)
	return err
}

func (p *PostgresStore) createUser(user *User) error {
	query := `insert into users
	(username,email,password)
	values ($1,$2,$3)`
	_, err := p.db.Query(query, user.Username,
		user.Email, user.Password)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStore) createItem(item *Item) error {
	query := `insert into items
	(creatorId,itemOwnerId,status,message)
	values ($1,$2,$3,$4)`
	_, err := p.db.Query(query, item.CreatorId,
		item.ItemOwnerId, item.Status, item.Message)

	if err != nil {
		return err
	}
	return nil
}
func (p *PostgresStore) deleteItem(id, userId int) error {
	query := `Update items
	Set isDeleted=TRUE
	Where id= $1 AND creatorId= $2`
	_, err := p.db.Query(query, id, userId)
	if err != nil {
		return err
	}
	return nil

}
func (p *PostgresStore) changeItemStatus(itemId int, status string, userId int) error {
	query := `Update items
			Set status= $1
			Where id= $2 AND creatorId= $3`

	_, err := p.db.Query(query, status, itemId, userId)
	if err != nil {
		return err
	}
	return nil
}
func (p *PostgresStore) getItemsByCreator(userId int) ([]*Item, error) {
	rows, err := p.db.Query(`select * from items where creatorId=$1 AND isDeleted=FALSE`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*Item{}

	for rows.Next() {
		item := Item{}
		if err := rows.Scan(&item.Id, &item.CreatorId,
			&item.ItemOwnerId, &item.Status, &item.Message, &item.IsDeleted); err != nil {
			return items, err
		}
		items = append(items, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
func (p *PostgresStore) getItemsofUser(ownerId int) ([]*Item, error) {
	rows, err := p.db.Query(`select * from items where itemOwnerId=$1`, ownerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*Item{}

	for rows.Next() {
		item := Item{}
		if err := rows.Scan(&item.Id, &item.CreatorId,
			&item.ItemOwnerId, &item.Status, &item.Message, &item.IsDeleted); err != nil {
			return items, err
		}
		items = append(items, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (p *PostgresStore) getAllUser() ([]*User, error) {
	rows, err := p.db.Query(`select * from users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		user := User{}
		if err := rows.Scan(&user.Id, &user.Username, &user.Password,
			&user.Email); err != nil {
			return users, err
		}
		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil

}
func (p *PostgresStore) getUserById(id int) (*User, error) {

	row := p.db.QueryRow(`SELECT * FROM users WHERE id=$1`, id)

	user := User{}
	if err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("User %d not found", id)
		}
		return nil, err
	}

	return &user, nil
}
func (p *PostgresStore) FindByEmail(email string) (*User, error) {

	user := User{}

	query := `SELECT id, username, password, email FROM users WHERE email=$1`

	row := p.db.QueryRow(query, email)
	if err := row.Scan(&user.Id, &user.Username, &user.Password, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

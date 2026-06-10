package audiobookshelf

import (
	"context"
)

// CreateCollectionRequest are the parameters for CreateCollection.
type CreateCollectionRequest struct {
	LibraryID   string   `json:"libraryId"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Books       []string `json:"books,omitempty"`
}

// UpdateCollectionRequest are the parameters for UpdateCollection.
// Nil/zero fields are left unchanged.
type UpdateCollectionRequest struct {
	LibraryID   string   `json:"libraryId,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Books       []string `json:"books,omitempty"`
}

func collectionPath(id string, rest ...string) string {
	return apiPath("collections").Seg(id).Lit(rest...).String()
}

// CreateCollection creates a collection (POST /api/collections).
func (c *Client) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*Collection, error) {
	var collection Collection
	if err := c.Post(ctx, apiPath("collections").String(), req, &collection); err != nil {
		return nil, err
	}

	collection.client = c

	return &collection, nil
}

// Collections returns all collections accessible to the user
// (GET /api/collections).
func (c *Client) Collections(ctx context.Context) ([]Collection, error) {
	var resp struct {
		Collections []Collection `json:"collections"`
	}

	if err := c.Get(ctx, apiPath("collections").String(), &resp); err != nil {
		return nil, err
	}

	for i := range resp.Collections {
		resp.Collections[i].client = c
	}

	return resp.Collections, nil
}

// Collection returns a collection (GET /api/collections/:id). include may
// be "rssfeed" or empty.
func (c *Client) Collection(ctx context.Context, id string, include string) (*Collection, error) {
	pb := apiPath("collections").Seg(id)
	if include != "" {
		pb.Set("include", include)
	}

	var collection Collection
	if err := c.Get(ctx, pb.String(), &collection); err != nil {
		return nil, err
	}

	collection.client = c

	return &collection, nil
}

// UpdateCollection updates a collection (PATCH /api/collections/:id).
func (c *Client) UpdateCollection(ctx context.Context, id string, req *UpdateCollectionRequest) (*Collection, error) {
	var collection Collection
	if err := c.Patch(ctx, collectionPath(id), req, &collection); err != nil {
		return nil, err
	}
	collection.client = c
	return &collection, nil
}

// DeleteCollection deletes a collection (DELETE /api/collections/:id).
func (c *Client) DeleteCollection(ctx context.Context, id string) error {
	return c.Delete(ctx, collectionPath(id), nil)
}

// AddBookToCollection adds a book library item to a collection
// (POST /api/collections/:id/book).
func (c *Client) AddBookToCollection(ctx context.Context, id, bookID string) (*Collection, error) {
	var collection Collection
	if err := c.Post(ctx, collectionPath(id, "book"), map[string]string{"id": bookID}, &collection); err != nil {
		return nil, err
	}

	collection.client = c

	return &collection, nil
}

// RemoveBookFromCollection removes a book library item from a collection
// (DELETE /api/collections/:id/book/:bookId).
func (c *Client) RemoveBookFromCollection(ctx context.Context, id, bookID string) (*Collection, error) {
	path := apiPath("collections").Seg(id).Lit("book").Seg(bookID).String()

	var collection Collection
	if err := c.Delete(ctx, path, &collection); err != nil {
		return nil, err
	}

	collection.client = c

	return &collection, nil
}

// BatchAddBooksToCollection adds multiple book library items to a
// collection (POST /api/collections/:id/batch/add).
func (c *Client) BatchAddBooksToCollection(ctx context.Context, id string, bookIDs []string) (*Collection, error) {
	var collection Collection
	if err := c.Post(ctx, collectionPath(id, "batch", "add"), map[string]any{"books": bookIDs}, &collection); err != nil {
		return nil, err
	}

	collection.client = c

	return &collection, nil
}

// BatchRemoveBooksFromCollection removes multiple book library items from
// a collection (POST /api/collections/:id/batch/remove).
func (c *Client) BatchRemoveBooksFromCollection(ctx context.Context, id string, bookIDs []string) (*Collection, error) {
	var collection Collection
	if err := c.Post(ctx, collectionPath(id, "batch", "remove"), map[string]any{"books": bookIDs}, &collection); err != nil {
		return nil, err
	}

	collection.client = c

	return &collection, nil
}

// Update updates the collection and refreshes its fields in place. See
// Client.UpdateCollection.
func (col *Collection) Update(ctx context.Context, req *UpdateCollectionRequest) error {
	updated, err := col.client.UpdateCollection(ctx, col.ID, req)
	if err != nil {
		return err
	}
	*col = *updated
	return nil
}

// Delete deletes the collection. See Client.DeleteCollection.
func (col *Collection) Delete(ctx context.Context) error {
	return col.client.DeleteCollection(ctx, col.ID)
}

// AddBook adds a book to the collection and refreshes its fields in
// place. See Client.AddBookToCollection.
func (col *Collection) AddBook(ctx context.Context, bookID string) error {
	updated, err := col.client.AddBookToCollection(ctx, col.ID, bookID)
	if err != nil {
		return err
	}
	*col = *updated
	return nil
}

// RemoveBook removes a book from the collection and refreshes its fields
// in place. See Client.RemoveBookFromCollection.
func (col *Collection) RemoveBook(ctx context.Context, bookID string) error {
	updated, err := col.client.RemoveBookFromCollection(ctx, col.ID, bookID)
	if err != nil {
		return err
	}
	*col = *updated
	return nil
}

// AddBooks adds multiple books to the collection and refreshes its
// fields in place. See Client.BatchAddBooksToCollection.
func (col *Collection) AddBooks(ctx context.Context, bookIDs []string) error {
	updated, err := col.client.BatchAddBooksToCollection(ctx, col.ID, bookIDs)
	if err != nil {
		return err
	}
	*col = *updated
	return nil
}

// RemoveBooks removes multiple books from the collection and refreshes
// its fields in place. See Client.BatchRemoveBooksFromCollection.
func (col *Collection) RemoveBooks(ctx context.Context, bookIDs []string) error {
	updated, err := col.client.BatchRemoveBooksFromCollection(ctx, col.ID, bookIDs)
	if err != nil {
		return err
	}
	*col = *updated
	return nil
}

package firestore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const endpoint = "https://firestore.googleapis.com/v1"

type Client struct {
	projectID string
	uid       string
	idToken   string
	http      *http.Client
}

func New(projectID, uid, idToken string) *Client {
	return &Client{projectID: projectID, uid: uid, idToken: idToken, http: http.DefaultClient}
}

type Document struct {
	ID     string
	Fields Fields
}

func (c *Client) databasePath() string {
	return fmt.Sprintf("projects/%s/databases/(default)/documents", c.projectID)
}

func (c *Client) userPath() string {
	return c.databasePath() + "/users/" + c.uid
}

func (c *Client) docName(collection, id string) string {
	return c.userPath() + "/" + collection + "/" + id
}

// List returns every document in the user's collection, ordered by the given field when not empty.
func (c *Client) List(ctx context.Context, collection, orderBy string) ([]Document, error) {
	query := map[string]any{"from": []map[string]any{{"collectionId": collection}}}
	if orderBy != "" {
		query["orderBy"] = []map[string]any{{"field": map[string]any{"fieldPath": orderBy}, "direction": "ASCENDING"}}
	}

	var resp []struct {
		Document *struct {
			Name   string `json:"name"`
			Fields Fields `json:"fields"`
		} `json:"document"`
	}
	if err := c.call(ctx, http.MethodPost, c.userPath()+":runQuery", nil, map[string]any{"structuredQuery": query}, &resp); err != nil {
		return nil, err
	}

	docs := make([]Document, 0, len(resp))
	for _, r := range resp {
		if r.Document == nil {
			continue
		}
		docs = append(docs, Document{ID: path.Base(r.Document.Name), Fields: r.Document.Fields})
	}
	return docs, nil
}

func (c *Client) Create(ctx context.Context, collection string, doc Document) error {
	q := url.Values{"documentId": {doc.ID}}
	return c.call(ctx, http.MethodPost, c.userPath()+"/"+collection, q, map[string]any{"fields": doc.Fields}, nil)
}

// Patch updates only the fields present in doc.Fields.
func (c *Client) Patch(ctx context.Context, collection string, doc Document) error {
	q := url.Values{}
	for key := range doc.Fields {
		q.Add("updateMask.fieldPaths", key)
	}
	return c.call(ctx, http.MethodPatch, c.docName(collection, doc.ID), q, map[string]any{"fields": doc.Fields}, nil)
}

func (c *Client) Delete(ctx context.Context, collection, id string) error {
	return c.call(ctx, http.MethodDelete, c.docName(collection, id), nil, nil, nil)
}

// Move atomically recreates the document in another collection and deletes the original.
func (c *Client) Move(ctx context.Context, from, to string, doc Document) error {
	writes := []map[string]any{
		{"update": map[string]any{"name": c.docName(to, doc.ID), "fields": doc.Fields}},
		{"delete": c.docName(from, doc.ID)},
	}
	return c.call(ctx, http.MethodPost, c.databasePath()+":commit", nil, map[string]any{"writes": writes}, nil)
}

func (c *Client) call(ctx context.Context, method, name string, query url.Values, body, out any) error {
	u := endpoint + "/" + name
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.idToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("firestore %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

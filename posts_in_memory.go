package main

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"slices"
// 	"strconv"
// 	"strings"
// 	"sync"
// )

// type post struct {
// 	ID      int    `json:"id"`
// 	Author  string `json:"author"`
// 	Message string `json:"message"`
// }

// var (
// 	posts = map[int]post{
// 		0: {ID: 0, Author: "Bilbo Beggins", Message: "Let the adventure begin..."},
// 		1: {ID: 1, Author: "Obi-Wan Kenobi", Message: "Hello there!"},
// 		2: {ID: 2, Author: "Geralt of Rivia", Message: "Hmmm... Wind's howling..."},
// 		3: {ID: 3, Author: "Obi-Wan Kenobi", Message: "May the force be with you ⚡"},
// 		4: {ID: 4, Author: "R2-D2", Message: "May the 4th bla-bla bee-boop"},
// 	}
// 	nextID = 5
// 	mu     sync.Mutex
// )

// func (app *application) getPostsList() ([]post, error) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	var postList []post
// 	for _, post := range posts {
// 		postList = append(postList, post)
// 	}

// 	slices.SortFunc(postList, func(a, b post) int {
// 		if a.ID == b.ID {
// 			return 0
// 		} else if a.ID < b.ID {
// 			return 1
// 		} else {
// 			return -1
// 		}
// 	})

// 	return postList, nil
// }

// func (app *application) getPosts(w http.ResponseWriter, r *http.Request) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	var postList []post
// 	for _, post := range posts {
// 		postList = append(postList, post)
// 	}

// 	w.Header().Set("Content-Type", "application/json")

// 	err := json.NewEncoder(w).Encode(postList)
// 	if err != nil {
// 		log.Printf("Failed to encode post to JSON: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
// }

// func (app *application) getPost(w http.ResponseWriter, r *http.Request) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	id := r.PathValue("id")
// 	postID, err := strconv.Atoi(id)
// 	if err != nil {
// 		log.Printf("Failed to convert id from query to int: %v", err)
// 		http.Error(w, "Invalid post id (id must be numeric)", http.StatusBadRequest)
// 		return
// 	}

// 	post, exists := posts[postID]
// 	if !exists {
// 		http.Error(w, "Post not found", http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")

// 	err = json.NewEncoder(w).Encode(post)
// 	if err != nil {
// 		log.Printf("Failed to encode post to JSON: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
// }

// func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	var newPost post
// 	if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
// 		log.Printf("Failed to parse payload: %v", err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	newPost.Author = strings.TrimSpace(newPost.Author)
// 	if newPost.Author == "" {
// 		http.Error(w, "Missing field: author", http.StatusBadRequest)
// 		return
// 	}

// 	newPost.Message = strings.TrimSpace(newPost.Message)
// 	if newPost.Message == "" {
// 		http.Error(w, "Missing field: message", http.StatusBadRequest)
// 		return
// 	}

// 	newPost.ID = nextID
// 	nextID++

// 	posts[newPost.ID] = newPost

// 	w.WriteHeader(http.StatusCreated)

// 	err := json.NewEncoder(w).Encode(newPost)
// 	if err != nil {
// 		log.Printf("Failed to encode post to JSON: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
// }

// func (app *application) updatePost(w http.ResponseWriter, r *http.Request) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	id := r.PathValue("id")
// 	postID, err := strconv.Atoi(id)
// 	if err != nil {
// 		log.Printf("Failed to convert id from query to int: %v", err)
// 		http.Error(w, "Invalid post id (id must be numeric)", http.StatusBadRequest)
// 		return
// 	}

// 	originalPost, exists := posts[postID]
// 	if !exists {
// 		http.Error(w, "Post not found", http.StatusNotFound)
// 		return
// 	}

// 	var updatedPost post
// 	if err := json.NewDecoder(r.Body).Decode(&updatedPost); err != nil {
// 		log.Printf("Failed to parse payload: %v", err)
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	updatedPost.Author = strings.TrimSpace(updatedPost.Author)
// 	updatedPost.Message = strings.TrimSpace(updatedPost.Message)

// 	if updatedPost.Author == "" && updatedPost.Message == "" {
// 		http.Error(w, "Bad request: no data provided", http.StatusBadRequest)
// 		return
// 	}

// 	updatedPost.ID = postID

// 	if updatedPost.Author == "" {
// 		updatedPost.Author = originalPost.Author
// 	}
// 	if updatedPost.Message == "" {
// 		updatedPost.Message = originalPost.Message
// 	}

// 	posts[postID] = updatedPost

// 	w.Header().Set("Content-Type", "application/json")

// 	err = json.NewEncoder(w).Encode(updatedPost)
// 	if err != nil {
// 		log.Printf("Failed to encode post to JSON: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}
// }

// func (app *application) deletePost(w http.ResponseWriter, r *http.Request) {
// 	mu.Lock()
// 	defer mu.Unlock()

// 	id := r.PathValue("id")
// 	postID, err := strconv.Atoi(id)
// 	if err != nil {
// 		log.Printf("Failed to convert id from query to int: %v", err)
// 		http.Error(w, "Invalid post id (id must be numeric)", http.StatusBadRequest)
// 		return
// 	}

// 	if _, exists := posts[postID]; !exists {
// 		http.Error(w, "Post not found", http.StatusNotFound)
// 		return
// 	}

// 	delete(posts, postID)
// 	w.WriteHeader(http.StatusNoContent)
// }

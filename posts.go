package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type post struct {
	ID      int    `json:"id"`
	Author  string `json:"author"`
	Message string `json:"message"`
}

var (
	posts = map[int]post{
		0: {ID: 0, Author: "Bilbo Beggins", Message: "Let the adventure begin..."},
		1: {ID: 1, Author: "Obi-Wan Kenobi", Message: "Hello there!"},
		2: {ID: 2, Author: "Geralt of Rivia", Message: "Hmmm... Wind's howling..."},
		3: {ID: 3, Author: "Obi-Wan Kenobi", Message: "May the force be with you ⚡"},
	}
	nextID = 4
	mu     sync.Mutex
)

func getPosts(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var postList []post
	for _, post := range posts {
		postList = append(postList, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postList)
}

func getPost(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	id := r.PathValue("id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to convert id from query to int:", err)
		http.Error(w, "Invalid post id (id must be numeric)", http.StatusBadRequest)
		return
	}

	post, exists := posts[postID]
	if !exists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var post post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		log.Println("Failed to parse payload:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post.Author = strings.TrimSpace(post.Author)
	post.Message = strings.TrimSpace(post.Message)

	if post.Author == "" || post.Message == "" {
		http.Error(w, "Bad request: not enough data provided", http.StatusBadRequest)
		return
	}

	post.ID = nextID
	nextID++

	posts[post.ID] = post
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func updatePost(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	id := r.PathValue("id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to convert id from query to int:", err)
		http.Error(w, "Invalid post id (id must be numeric)", http.StatusBadRequest)
		return
	}

	originalPost, exists := posts[postID]
	if !exists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	var updatedPost post
	if err := json.NewDecoder(r.Body).Decode(&updatedPost); err != nil {
		log.Println("Failed to parse payload:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedPost.Author = strings.TrimSpace(updatedPost.Author)
	updatedPost.Message = strings.TrimSpace(updatedPost.Message)

	if updatedPost.Author == "" && updatedPost.Message == "" {
		http.Error(w, "Bad request: no data provided", http.StatusBadRequest)
		return
	}

	updatedPost.ID = postID

	if updatedPost.Author == "" {
		updatedPost.Author = originalPost.Author
	}
	if updatedPost.Message == "" {
		updatedPost.Message = originalPost.Message
	}

	posts[postID] = updatedPost
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPost)
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	id := r.PathValue("id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Failed to convert id from query to int:", err)
		http.Error(w, "Invalid post id (id must be numeric)", http.StatusBadRequest)
		return
	}

	if _, exists := posts[postID]; !exists {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	delete(posts, postID)
	w.WriteHeader(http.StatusNoContent)
}

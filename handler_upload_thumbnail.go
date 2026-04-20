package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	vidMetaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find video", err)
		return
	}

	if vidMetaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Video user ID does not match token user ID", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")

	if !strings.HasPrefix(mediaType, "image/") {
		respondWithError(w, http.StatusBadRequest, "Unable to parse file prefix. Not an image.", nil)
		return
	}

	extension := strings.TrimSpace(strings.TrimPrefix(mediaType, "image/"))
	filePath := filepath.Join(cfg.assetsRoot, videoID.String()+"."+extension)

	createdFile, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Couldn't create file: %v", err)
		return
	}

	_, err = io.Copy(createdFile, file)
	if err != nil {
		log.Fatalf("Couldn't write file: %v", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, videoID.String(), extension)
	vidMetaData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(vidMetaData)
	if err != nil {
		log.Fatalf("Couldn't update video: %v", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vidMetaData)
}

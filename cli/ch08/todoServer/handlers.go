package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidData = errors.New("invalid data")
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	content := "There's an API here"

	replyTextContent(w, r, http.StatusOK, content)
}

func getAllHandler(w http.ResponseWriter, r *http.Request) {
	resp := &todoResponse{
		Results: list,
	}

	replyJSON(w, r, http.StatusOK, resp)
}

func getOneHandler(w http.ResponseWriter, r *http.Request) {
	id, err := validateID(r.PathValue("id"))
	if err != nil {
		replyError(w, r, http.StatusBadRequest, err.Error())
	}

	resp := &todoResponse{
		Results: list[id-1 : id],
	}

	replyJSON(w, r, http.StatusOK, resp)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := validateID(r.PathValue("id"))
	if err != nil {
		replyError(w, r, http.StatusBadRequest, err.Error())
	}

	list.Delete(id)
	if err := list.Save(todoFile); err != nil {
		replyError(w, r, http.StatusAccepted)
	}
}

func validateID(pathId string) (int, error) {
	id, err := strconv.Atoi(pathId)
	if err != nil {
		return 0, fmt.Errorf("%w: Invalid ID: %s", ErrInvalidData, err)
	}

	if id < 1 {
		return 0, fmt.Errorf("%w, Invalid ID: Less than one", ErrInvalidData)
	}

	if id > len(list) {
		return 0, fmt.Errorf("%w, ID: %d not found", ErrInvalidData, id)
	}

	return id, nil
}

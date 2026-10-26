package documentshttp

import (
	"net/http"
)

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}

const (
	foldersPath  = "/api/folders"
	folderPath   = "/api/folders/{folder_id}"
	itemsPath    = "/api/documents"
	itemPath     = "/api/documents/{document_id}"
	uploadsPath  = "/api/document-uploads"
	uploadPath   = "/api/document-uploads/{upload_id}"
	archivesPath = "/api/document-archives"
	archivePath  = "/api/document-archives/{archive_id}"
)

func (handler *Handler) Routes() []Route {
	if handler == nil {
		return nil
	}
	return []Route{
		{Method: http.MethodGet, Pattern: foldersPath,
			Handler: http.HandlerFunc(handler.listFolders)},
		{Method: http.MethodPost, Pattern: foldersPath,
			Handler: http.HandlerFunc(handler.createFolder)},
		{Method: http.MethodGet, Pattern: folderPath,
			Handler: http.HandlerFunc(handler.getFolder)},
		{Method: http.MethodPut, Pattern: folderPath,
			Handler: http.HandlerFunc(handler.updateFolder)},
		{Method: http.MethodDelete, Pattern: folderPath,
			Handler: http.HandlerFunc(handler.deleteFolder)},
		{Method: http.MethodPost, Pattern: folderPath + "/move",
			Handler: http.HandlerFunc(handler.moveFolder)},
		{Method: http.MethodPost, Pattern: folderPath + "/restore",
			Handler: http.HandlerFunc(handler.restoreFolder)},
		{Method: http.MethodGet, Pattern: folderPath + "/contents",
			Handler: http.HandlerFunc(handler.listFolderContents)},
		{Method: http.MethodGet, Pattern: folderPath + "/download",
			Handler: http.HandlerFunc(handler.legacyFolderArchive)},
		{Method: http.MethodGet, Pattern: itemsPath,
			Handler: http.HandlerFunc(handler.listItems)},
		{Method: http.MethodPost, Pattern: itemsPath,
			Handler: http.HandlerFunc(handler.legacyCreateDocument)},
		{Method: http.MethodPost, Pattern: itemsPath + "/batch",
			Handler: http.HandlerFunc(handler.batchItems)},
		{Method: http.MethodGet, Pattern: itemPath,
			Handler: http.HandlerFunc(handler.getItem)},
		{Method: http.MethodPut, Pattern: itemPath,
			Handler: http.HandlerFunc(handler.updateItem)},
		{Method: http.MethodDelete, Pattern: itemPath,
			Handler: http.HandlerFunc(handler.deleteItem)},
		{Method: http.MethodPost, Pattern: itemPath + "/move",
			Handler: http.HandlerFunc(handler.moveItem)},
		{Method: http.MethodPost, Pattern: itemPath + "/copy",
			Handler: http.HandlerFunc(handler.copyItem)},
		{Method: http.MethodPost, Pattern: itemPath + "/restore",
			Handler: http.HandlerFunc(handler.restoreItem)},
		{Method: http.MethodPost, Pattern: itemPath + "/purge",
			Handler: http.HandlerFunc(handler.purgeItem)},
		{Method: http.MethodPost, Pattern: itemPath + "/purge/cancel",
			Handler: http.HandlerFunc(handler.cancelItemPurge)},
		{Method: http.MethodGet, Pattern: itemPath + "/download",
			Handler: http.HandlerFunc(handler.downloadItem)},
		{Method: http.MethodGet, Pattern: itemPath + "/versions",
			Handler: http.HandlerFunc(handler.listVersions)},
		{Method: http.MethodPost, Pattern: itemPath + "/versions",
			Handler: http.HandlerFunc(handler.legacyCreateVersion)},
		{Method: http.MethodGet, Pattern: itemPath + "/versions/latest",
			Handler: http.HandlerFunc(handler.latestVersion)},
		{Method: http.MethodGet,
			Pattern: itemPath + "/versions/{version_id}/content",
			Handler: http.HandlerFunc(handler.downloadVersion)},
		{Method: http.MethodPost, Pattern: uploadsPath,
			Handler: http.HandlerFunc(handler.createUpload)},
		{Method: http.MethodGet, Pattern: uploadPath,
			Handler: http.HandlerFunc(handler.getUpload)},
		{Method: http.MethodPut, Pattern: uploadPath + "/content",
			Handler: http.HandlerFunc(handler.putUploadContent)},
		{Method: http.MethodPost, Pattern: uploadPath + "/cancel",
			Handler: http.HandlerFunc(handler.cancelUpload)},
		{Method: http.MethodPost, Pattern: archivesPath,
			Handler: http.HandlerFunc(handler.createArchive)},
		{Method: http.MethodGet, Pattern: archivePath,
			Handler: http.HandlerFunc(handler.getArchive)},
		{Method: http.MethodGet, Pattern: archivePath + "/content",
			Handler: http.HandlerFunc(handler.downloadArchive)},
	}
}

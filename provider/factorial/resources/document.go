// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// Document fetches Factorial HR documents.
type Document struct {
	client Client
}

// NewDocument creates a new Document resource.
func NewDocument(client Client) *Document {
	return &Document{client: client}
}

// Name returns the resource type name.
func (r *Document) Name() string {
	return "factorial_document"
}

// Fetch retrieves all Factorial HR documents.
func (r *Document) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	documents, err := r.client.FetchPaginated(ctx, "resources/documents/documents")
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(documents))

	for _, doc := range documents {
		resource := make(core.Resource)

		// Basic info
		if id, ok := doc["id"].(float64); ok {
			resource["id"] = int64(id)
		}
		if employeeID, ok := doc["employee_id"].(float64); ok {
			resource["employee_id"] = int64(employeeID)
		}
		if folderID, ok := doc["folder_id"].(float64); ok {
			resource["folder_id"] = int64(folderID)
		}

		resource["name"] = doc["name"]
		resource["filename"] = doc["filename"]
		resource["public"] = doc["public"]
		resource["is_visible"] = doc["is_visible"]

		// Signature info
		resource["signed"] = doc["signed"]
		resource["pending_signature"] = doc["pending_signature"]
		resource["signature_request_id"] = doc["signature_request_id"]

		// File info
		resource["file_url"] = doc["file_url"]
		resource["file_size"] = doc["file_size"]
		resource["content_type"] = doc["content_type"]

		// Timestamps
		resource["created_at"] = doc["created_at"]
		resource["updated_at"] = doc["updated_at"]

		// Computed fields for policy evaluation
		var isSigned, isPending bool
		if v, ok := doc["signed"].(bool); ok {
			isSigned = v
		}
		if v, ok := doc["pending_signature"].(bool); ok {
			isPending = v
		}

		resource["is_signed"] = isSigned
		resource["is_pending"] = isPending
		resource["needs_signature"] = !isSigned && isPending

		// Check if public document
		var isPublic bool
		if v, ok := doc["public"].(bool); ok {
			isPublic = v
		}
		resource["is_public"] = isPublic

		// Check document name for common compliance documents
		var name string
		if v, ok := doc["name"].(string); ok {
			name = v
		}
		nameLower := strings.ToLower(name)
		resource["is_nda"] = strings.Contains(nameLower, "nda") || strings.Contains(nameLower, "non-disclosure") || strings.Contains(nameLower, "confidentiality")
		resource["is_contract"] = strings.Contains(nameLower, "contract") || strings.Contains(nameLower, "agreement")
		resource["is_policy"] = strings.Contains(nameLower, "policy") || strings.Contains(nameLower, "handbook")
		resource["is_training"] = strings.Contains(nameLower, "training") || strings.Contains(nameLower, "certification")

		// Check visibility
		var isVisible bool
		if v, ok := doc["is_visible"].(bool); ok {
			isVisible = v
		}
		resource["is_visible_to_employee"] = isVisible

		resources = append(resources, resource)
	}

	return resources, nil
}

// Folder fetches Factorial HR document folders.
type Folder struct {
	client Client
}

// NewFolder creates a new Folder resource.
func NewFolder(client Client) *Folder {
	return &Folder{client: client}
}

// Name returns the resource type name.
func (r *Folder) Name() string {
	return "factorial_folder"
}

// Fetch retrieves all Factorial HR document folders.
func (r *Folder) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	folders, err := r.client.FetchPaginated(ctx, "resources/documents/folders")
	if err != nil {
		return nil, err
	}

	// Count documents per folder
	documents, err := r.client.FetchPaginated(ctx, "resources/documents/documents")
	folderDocCounts := make(map[float64]int)
	if err == nil {
		for _, doc := range documents {
			if folderID, ok := doc["folder_id"].(float64); ok {
				folderDocCounts[folderID]++
			}
		}
	}

	resources := make([]core.Resource, 0, len(folders))

	for _, folder := range folders {
		resource := make(core.Resource)

		// Basic info
		var folderID float64
		if id, ok := folder["id"].(float64); ok {
			resource["id"] = int64(id)
			folderID = id
		}
		if companyID, ok := folder["company_id"].(float64); ok {
			resource["company_id"] = int64(companyID)
		}

		resource["name"] = folder["name"]
		resource["active"] = folder["active"]

		// Timestamps
		resource["created_at"] = folder["created_at"]
		resource["updated_at"] = folder["updated_at"]

		// Computed fields for policy evaluation
		var folderName string
		if v, ok := folder["name"].(string); ok {
			folderName = v
		}
		resource["has_name"] = folderName != ""

		var isActive bool
		if v, ok := folder["active"].(bool); ok {
			isActive = v
		}
		resource["is_active"] = isActive

		resource["document_count"] = folderDocCounts[folderID]
		resource["has_documents"] = folderDocCounts[folderID] > 0
		resource["is_empty"] = folderDocCounts[folderID] == 0

		resources = append(resources, resource)
	}

	return resources, nil
}

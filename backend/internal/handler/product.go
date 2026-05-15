package handler

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/pkg/storage"
	"github.com/apsferreira/storemaker/internal/repository"
)

// trackProductPublished emite product.published quando is_active=true.
func trackProductPublished(ctx context.Context, userID, storeID, productID string) {
	if pulseClient == nil {
		return
	}
	pulseClient.Track(ctx, "product.published", userID, map[string]string{
		"store_id":   storeID,
		"product_id": productID,
	})
}

const maxPhotoSize = 5 * 1024 * 1024 // 5MB

var allowedMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// storageClient é o cliente MinIO injetado via InitStorage.
// É obrigatório — o servidor não sobe sem ele.
var storageClient *storage.Client

// InitStorage injeta o cliente de armazenamento nos handlers de produto.
// Deve ser chamado em main.go antes de registrar as rotas.
func InitStorage(c *storage.Client) {
	storageClient = c
}

func CreateProduct(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req model.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	if err := validateCreateProduct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	product, err := repository.CreateProduct(storeID, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "slug já existe nesta loja"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao criar produto"})
	}

	// BKL-1250: Pulse tracking — emite product.published se produto criado ativo.
	if product.IsActive {
		trackProductPublished(c.Context(), extractUserID(c), storeID, product.ID)
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

func GetProduct(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	product, err := repository.GetProduct(storeID, c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar produto"})
	}
	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "produto não encontrado"})
	}

	return c.JSON(product)
}

func ListProducts(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	filter := buildFilter(c)
	products, total, err := repository.ListProducts(storeID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar produtos"})
	}

	if products == nil {
		products = []model.Product{}
	}

	totalPages := total / filter.PerPage
	if total%filter.PerPage != 0 {
		totalPages++
	}

	return c.JSON(model.PaginatedResponse{
		Data:       products,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	})
}

func UpdateProduct(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req model.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" || len(trimmed) > 255 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name inválido"})
		}
		req.Name = &trimmed
	}
	if req.PriceCents != nil && *req.PriceCents < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "price_cents deve ser >= 0"})
	}

	product, err := repository.UpdateProduct(storeID, c.Params("id"), req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "slug já existe nesta loja"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao atualizar produto"})
	}
	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "produto não encontrado"})
	}

	// BKL-1250: Pulse tracking — emite product.published quando is_active ativado explicitamente.
	if req.IsActive != nil && *req.IsActive && product.IsActive {
		trackProductPublished(c.Context(), extractUserID(c), storeID, product.ID)
	}

	return c.JSON(product)
}

func DeleteProduct(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	if err := repository.DeleteProduct(storeID, c.Params("id")); err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "produto não encontrado"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao deletar produto"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func ReorderProducts(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	var req model.ReorderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "corpo inválido"})
	}

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "items é obrigatório"})
	}

	if err := repository.ReorderProducts(storeID, req.Items); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao reordenar produtos"})
	}

	return c.JSON(fiber.Map{"message": "produtos reordenados"})
}

func UploadProductPhotos(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	productID := c.Params("id")
	// Verify product belongs to store
	product, err := repository.GetProduct(storeID, productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao verificar produto"})
	}
	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "produto não encontrado"})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "form-data inválido"})
	}

	files := form.File["photos"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nenhuma foto enviada"})
	}

	var photos []model.ProductPhoto
	for i, file := range files {
		if err := validateUploadFile(file); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("arquivo %s: %s", file.Filename, err.Error())})
		}

		ext := strings.ToLower(strings.TrimPrefix(getFileExt(file.Filename), "."))
		if ext == "" {
			ext = "jpg"
		}
		objectName := fmt.Sprintf("stores/%s/products/%s/%s.%s", storeID, productID, uuid.New().String(), ext)

		f, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao abrir foto para upload"})
		}

		contentType := detectContentType(file)
		url, err := storageClient.Upload(context.Background(), objectName, f, file.Size, contentType)
		f.Close()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao enviar foto para armazenamento"})
		}

		photo, err := repository.CreatePhoto(productID, url, i)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao registrar foto"})
		}
		photos = append(photos, *photo)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": photos})
}

func ImportProductsCSV(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "arquivo CSV não fornecido"})
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "apenas arquivos .csv são aceitos"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao abrir arquivo"})
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "erro ao ler cabeçalho CSV"})
	}

	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	required := []string{"name", "slug", "price_cents"}
	for _, r := range required {
		if _, ok := colMap[r]; !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("coluna obrigatória ausente: %s", r),
			})
		}
	}

	var created int
	var errors []string
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		lineNum++
		if err != nil {
			errors = append(errors, fmt.Sprintf("linha %d: erro de leitura", lineNum))
			continue
		}

		name := strings.TrimSpace(record[colMap["name"]])
		slug := strings.TrimSpace(record[colMap["slug"]])
		priceStr := strings.TrimSpace(record[colMap["price_cents"]])

		if name == "" || slug == "" || priceStr == "" {
			errors = append(errors, fmt.Sprintf("linha %d: campos obrigatórios vazios", lineNum))
			continue
		}

		priceCents, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("linha %d: price_cents inválido", lineNum))
			continue
		}

		req := model.CreateProductRequest{
			Name:                name,
			Slug:                slug,
			PriceCents:          priceCents,
			StockAlertThreshold: 5,
		}

		// Optional fields
		if idx, ok := colMap["description"]; ok && idx < len(record) {
			desc := strings.TrimSpace(record[idx])
			if desc != "" {
				req.Description = &desc
			}
		}
		if idx, ok := colMap["sku"]; ok && idx < len(record) {
			sku := strings.TrimSpace(record[idx])
			if sku != "" {
				req.SKU = &sku
			}
		}
		if idx, ok := colMap["stock_quantity"]; ok && idx < len(record) {
			sq, err := strconv.Atoi(strings.TrimSpace(record[idx]))
			if err == nil {
				req.StockQuantity = sq
			}
		}
		if idx, ok := colMap["categoria_id"]; ok && idx < len(record) {
			catID := strings.TrimSpace(record[idx])
			if catID != "" {
				req.CategoriaID = &catID
			}
		}

		_, err = repository.CreateProduct(storeID, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("linha %d: %s", lineNum, err.Error()))
			continue
		}
		created++
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"imported": created,
		"errors":   errors,
	})
}

func GetLowStockAlert(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	products, err := repository.GetLowStockProducts(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar estoque baixo"})
	}

	if products == nil {
		products = []model.Product{}
	}

	return c.JSON(fiber.Map{"data": products})
}

// Public catalog - no auth
func PublicCatalog(c *fiber.Ctx) error {
	storeID := c.Query("store_id")
	if storeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "store_id é obrigatório"})
	}

	filter := buildFilter(c)
	products, total, err := repository.ListPublicCatalog(storeID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar catálogo"})
	}

	if products == nil {
		products = []model.Product{}
	}

	// Load photos for each product
	for i := range products {
		photos, err := repository.GetProductPhotos(products[i].ID)
		if err == nil {
			products[i].Photos = photos
		}
		variations, err := repository.GetProductVariations(products[i].ID)
		if err == nil {
			products[i].Variations = variations
		}
	}

	totalPages := total / filter.PerPage
	if total%filter.PerPage != 0 {
		totalPages++
	}

	return c.JSON(model.PaginatedResponse{
		Data:       products,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	})
}

// --- helpers ---

func validateCreateProduct(req model.CreateProductRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)

	if req.Name == "" || req.Slug == "" {
		return fmt.Errorf("name e slug são obrigatórios")
	}
	if len(req.Name) > 255 || len(req.Slug) > 255 {
		return fmt.Errorf("name e slug devem ter até 255 caracteres")
	}
	if req.PriceCents < 0 {
		return fmt.Errorf("price_cents deve ser >= 0")
	}
	if req.StockQuantity < 0 {
		return fmt.Errorf("stock_quantity deve ser >= 0")
	}

	for _, v := range req.Variations {
		if strings.TrimSpace(v.Name) == "" || strings.TrimSpace(v.Value) == "" {
			return fmt.Errorf("variações devem ter name e value")
		}
	}

	return nil
}

func validateUploadFile(file *multipart.FileHeader) error {
	if file.Size > maxPhotoSize {
		return fmt.Errorf("tamanho excede limite de 5MB")
	}

	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo")
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("erro ao ler arquivo")
	}

	mimeType := http.DetectContentType(buf[:n])
	if !allowedMIME[mimeType] {
		return fmt.Errorf("tipo de arquivo não permitido: %s (aceitos: jpeg, png, webp, gif)", mimeType)
	}

	return nil
}

// getFileExt retorna a extensão do arquivo sem o ponto (ex: "jpg").
// Usa apenas os caracteres após o último "." no nome do arquivo.
func getFileExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i+1:]
		}
	}
	return ""
}

// detectContentType lê os primeiros 512 bytes do arquivo para detectar o MIME type.
// Retorna "application/octet-stream" em caso de erro.
func detectContentType(file *multipart.FileHeader) string {
	f, err := file.Open()
	if err != nil {
		return "application/octet-stream"
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "application/octet-stream"
	}

	return http.DetectContentType(buf[:n])
}

func buildFilter(c *fiber.Ctx) model.ProductListFilter {
	page, perPage := paginationParams(c)

	filter := model.ProductListFilter{
		Page:    page,
		PerPage: perPage,
	}

	if catID := c.Query("categoria_id"); catID != "" {
		filter.CategoriaID = &catID
	}
	if minPrice := c.Query("min_price"); minPrice != "" {
		if v, err := strconv.ParseInt(minPrice, 10, 64); err == nil {
			filter.MinPrice = &v
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if v, err := strconv.ParseInt(maxPrice, 10, 64); err == nil {
			filter.MaxPrice = &v
		}
	}
	if inStock := c.Query("in_stock"); inStock == "true" {
		b := true
		filter.InStock = &b
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	return filter
}

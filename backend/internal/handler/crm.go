package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/apsferreira/storemaker/internal/model"
	"github.com/apsferreira/storemaker/internal/repository"
)

// dateRegex valida formato YYYY-MM-DD
var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// ListCustomers lista clientes da loja com total gasto, última compra e num pedidos.
// GET /admin/customers?search=&page=&per_page=
func ListCustomers(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	page, perPage := paginationParams(c)
	filter := model.CustomerListFilter{
		Page:    page,
		PerPage: perPage,
	}

	if search := c.Query("search"); search != "" {
		s := strings.TrimSpace(search)
		if len(s) > 255 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "search muito longo (max 255)"})
		}
		filter.Search = &s
	}

	customers, total, err := repository.ListCustomers(storeID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao listar clientes"})
	}

	if customers == nil {
		customers = []model.Customer{}
	}

	totalPages := total / filter.PerPage
	if total%filter.PerPage != 0 {
		totalPages++
	}

	return c.JSON(model.PaginatedResponse{
		Data:       customers,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	})
}

// GetCustomerDetail retorna detalhe do cliente com histórico de pedidos.
// GET /admin/customers/:id
func GetCustomerDetail(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	customerID := c.Params("id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id é obrigatório"})
	}

	detail, err := repository.GetCustomerDetail(storeID, customerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar cliente"})
	}
	if detail == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "cliente não encontrado"})
	}

	if detail.Orders == nil {
		detail.Orders = []model.Order{}
	}

	return c.JSON(detail)
}

// GetDashboard retorna dados do dashboard admin.
// GET /admin/dashboard
func GetDashboard(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	data, err := repository.GetDashboard(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao carregar dashboard"})
	}

	return c.JSON(data)
}

// GetSalesReport retorna relatório de vendas por período.
// GET /admin/reports/sales?from=YYYY-MM-DD&to=YYYY-MM-DD
func GetSalesReport(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "parâmetros from e to são obrigatórios (YYYY-MM-DD)"})
	}

	if !dateRegex.MatchString(from) || !dateRegex.MatchString(to) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "formato de data inválido (use YYYY-MM-DD)"})
	}

	if from > to {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "from deve ser anterior ou igual a to"})
	}

	report, err := repository.GetSalesReport(storeID, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao gerar relatório"})
	}

	return c.JSON(report)
}

// GetTopProducts retorna top 5 produtos por vendas e receita.
// GET /admin/reports/products
func GetTopProducts(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	products, err := repository.GetTopProducts(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar top produtos"})
	}

	return c.JSON(fiber.Map{"data": products})
}

// GetStockAlerts retorna produtos com estoque abaixo do threshold.
// GET /admin/reports/stock-alerts
func GetStockAlerts(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	alerts, err := repository.GetStockAlerts(storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao buscar alertas de estoque"})
	}

	return c.JSON(fiber.Map{"data": alerts})
}

// ExportSalesCSV exporta vendas por período em CSV.
// GET /admin/reports/export?format=csv&from=YYYY-MM-DD&to=YYYY-MM-DD
func ExportSalesCSV(c *fiber.Ctx) error {
	storeID, err := extractStoreID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	format := c.Query("format", "csv")
	if format != "csv" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "formato suportado: csv"})
	}

	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "parâmetros from e to são obrigatórios (YYYY-MM-DD)"})
	}

	if !dateRegex.MatchString(from) || !dateRegex.MatchString(to) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "formato de data inválido (use YYYY-MM-DD)"})
	}

	if from > to {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "from deve ser anterior ou igual a to"})
	}

	records, err := repository.GetSalesForExport(storeID, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao gerar export"})
	}

	filename := fmt.Sprintf("vendas_%s_%s.csv", from, to)
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	var buf bytes.Buffer
	// BOM para Excel reconhecer UTF-8
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(records); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao gerar CSV"})
	}
	writer.Flush()

	return c.Send(buf.Bytes())
}

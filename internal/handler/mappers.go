package handler

import (
	"github.com/product-catalog-service/internal/query"
	pb "github.com/product-catalog-service/proto/product/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MapProductResponseToProto maps an application response to a proto response.
func MapProductResponseToProto(resp *query.ProductResponse) *pb.Product {
	if resp == nil {
		return nil
	}

	product := &pb.Product{
		Id:          resp.ID,
		Name:        resp.Name,
		Description: resp.Description,
		Category:    resp.Category,
		BasePrice: &pb.Money{
			Numerator:   resp.BasePriceNumerator,
			Denominator: resp.BasePriceDenominator,
		},
		EffectivePrice: &pb.Money{
			Numerator:   resp.EffectivePriceNumerator,
			Denominator: resp.EffectivePriceDenominator,
		},
		HasActiveDiscount: resp.HasActiveDiscount,
		Status:            resp.Status,
		CreatedAt:         timestamppb.New(resp.CreatedAt),
		UpdatedAt:         timestamppb.New(resp.UpdatedAt),
	}

	if resp.DiscountPercent != nil {
		product.Discount = &pb.Discount{
			Percentage: *resp.DiscountPercent,
		}
		if resp.DiscountStartDate != nil {
			product.Discount.StartDate = timestamppb.New(*resp.DiscountStartDate)
		}
		if resp.DiscountEndDate != nil {
			product.Discount.EndDate = timestamppb.New(*resp.DiscountEndDate)
		}
	}

	return product
}

// MapListProductsResponseToProto maps an application response to a proto response.
func MapListProductsResponseToProto(resp *query.ListProductsResponse) *pb.ListProductsReply {
	if resp == nil {
		return &pb.ListProductsReply{}
	}

	products := make([]*pb.ProductSummary, len(resp.Products))
	for i, p := range resp.Products {
		summary := &pb.ProductSummary{
			Id:       p.ID,
			Name:     p.Name,
			Category: p.Category,
			BasePrice: &pb.Money{
				Numerator:   p.BasePriceNumerator,
				Denominator: p.BasePriceDenominator,
			},
			EffectivePrice: &pb.Money{
				Numerator:   p.EffectivePriceNumerator,
				Denominator: p.EffectivePriceDenominator,
			},
			HasActiveDiscount: p.HasActiveDiscount,
			Status:            p.Status,
			CreatedAt:         timestamppb.New(p.CreatedAt),
		}
		if p.DiscountPercent != nil {
			summary.DiscountPercent = *p.DiscountPercent
		}
		products[i] = summary
	}

	return &pb.ListProductsReply{
		Products:      products,
		NextPageToken: resp.NextPageToken,
		TotalCount:    resp.TotalCount,
	}
}

package mongodb

import (
	"reflect"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/v2/bson" // Single import for all BSON types
)

type DecimalCodec struct{}

func (dc *DecimalCodec) EncodeValue(ec bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
	d := val.Interface().(decimal.Decimal)
	// Convert decimal.Decimal to BSON Decimal128
	d128, err := bson.ParseDecimal128(d.String())
	if err != nil {
		return err
	}
	return vw.WriteDecimal128(d128)
}

func (dc *DecimalCodec) DecodeValue(dcx bson.DecodeContext, vr bson.ValueReader, val reflect.Value) error {
	d128, err := vr.ReadDecimal128()
	if err != nil {
		return err
	}
	d, err := decimal.NewFromString(d128.String())
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(d))
	return nil
}

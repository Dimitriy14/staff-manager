package elasticsearch

//import (
//	"context"
//
//	"github.com/Dimitriy14/staff-manager/logger"
//	transactionID "github.com/Dimitriy14/staff-manager/logger/transaction-id"
//
//	"github.com/olivere/elastic/v7"
//)
//
//const afterCallBackID = "ecdbd32b-7c24-4470-b468-5101e721c665"
//
//// AfterCallback is called after each request is processed by elasticsearch
//func AfterCallback(log logger.Logger) func(_ int64, request []elastic.BulkableRequest, resp *elastic.BulkResponse, err error) {
//	return func(_ int64, request []elastic.BulkableRequest, resp *elastic.BulkResponse, err error) {
//		ctx := transactionID.AddIDContext(context.Background(), afterCallBackID)
//		if err != nil {
//			log.Errorf(transactionID.FromContext(ctx), "Error persisting message %v to elasticsearch: %v", err)
//		}
//		//log.Debugf(transactionID.FromContext(ctx), "%s Made requests - (%v), is errors in response (%v)", prefix, request, resp.Errors)
//
//		if resp != nil && resp.Errors {
//			logErrorMsg(ctx, log, resp.Items)
//		}
//	}
//}
//
//func logErrorMsg(ctx context.Context, log logger.Logger, items []map[string]*elastic.BulkResponseItem) {
//	for _, item := range items {
//		for key, value := range item {
//			if value.Error != nil {
//				log.Errorf(transactionID.FromContext(ctx), "Error persisting message to elasticsearch. %s failed for %s/%s/%s. Error status: %d, type: %s, reason: %s", key, value.Index, value.Type, value.Id,
//					value.Status, value.Error.Type, value.Error.Reason,
//				)
//			}
//		}
//	}
//}

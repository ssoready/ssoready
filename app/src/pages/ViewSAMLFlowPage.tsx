import { useParams } from "react-router";
import { useQuery } from "@connectrpc/connect-query";
import {
  getSAMLConnection,
  getSAMLFlow,
  listSAMLFlows,
} from "@/gen/ssoready/v1/ssoready-SSOReadyService_connectquery";
import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import moment from "moment";
import formatXml from "xml-formatter";

export function ViewSAMLFlowPage() {
  const { samlFlowId } = useParams();
  const { data: samlFlow } = useQuery(getSAMLFlow, {
    id: samlFlowId,
  });

  return (
    <div>
      <Card>
        <CardHeader>
          <CardTitle>{samlFlowId}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex justify-between">
            <div>Started</div>
            <div>
              {samlFlow?.createTime &&
                moment(samlFlow.createTime.toDate()).fromNow()}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <code>{samlFlow?.authRedirectUrl}</code>
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <code>
            <pre>
              {samlFlow?.initiateRequest && formatXml(samlFlow.initiateRequest)}
            </pre>
          </code>
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <code>
            <pre>{samlFlow?.assertion && formatXml(samlFlow.assertion)}</pre>
          </code>

          <code>{samlFlow?.appRedirectUrl}</code>
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          Reedeemed at{" "}
          {samlFlow?.redeemTime &&
            moment(samlFlow.redeemTime.toDate()).format()}
        </CardContent>
      </Card>
    </div>
  );
}

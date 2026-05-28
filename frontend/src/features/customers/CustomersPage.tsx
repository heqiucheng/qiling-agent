import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { IntentBadge } from "../../components/customer/IntentBadge";
import { StageBadge } from "../../components/customer/StageBadge";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { listCustomers } from "../../lib/api/customers";
import type { Customer } from "../../types/customer";

export function CustomersPage() {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    let active = true;

    void listCustomers()
      .then((result) => {
        if (active) {
          setCustomers(result.items);
          setLoaded(true);
        }
      })
      .catch(() => {
        if (active) {
          setCustomers([]);
          setLoaded(true);
        }
      });

    return () => {
      active = false;
    };
  }, []);

  return (
    <section className="page">
      <header className="page__header">
        <div>
          <h1 className="page__title">客户</h1>
          <p className="page__subtitle">查看客户阶段、意向、主要顾虑和待跟进状态。</p>
        </div>
      </header>
      {customers.length > 0 ? (
        <Card>
          <div className="data-table" role="table" aria-label="客户列表">
            <div className="data-table__row data-table__row--head" role="row">
              <span role="columnheader">客户</span>
              <span role="columnheader">阶段</span>
              <span role="columnheader">意向</span>
              <span role="columnheader">负责人</span>
              <span role="columnheader">主要顾虑</span>
              <span role="columnheader">待办</span>
            </div>
            {customers.map((customer) => (
              <div className="data-table__row" role="row" key={customer.id}>
                <strong role="cell"><Link className="text-link" to={`/app/customers/${customer.id}`}>{customer.name}</Link></strong>
                <span role="cell"><StageBadge stage={customer.stage} /></span>
                <span role="cell"><IntentBadge level={customer.intent} /></span>
                <span role="cell">{customer.owner}</span>
                <span role="cell">{customer.concerns.join(" / ")}</span>
                <span role="cell">{customer.pendingTasks}</span>
              </div>
            ))}
          </div>
        </Card>
      ) : (
        <EmptyState
          title={loaded ? "暂无客户数据" : "正在读取客户数据"}
          message={loaded ? "可以先从数据接入页导入聊天记录。" : "如果后端未启动，将保持空状态。"}
          action="导入聊天记录后自动生成客户画像"
        />
      )}
    </section>
  );
}

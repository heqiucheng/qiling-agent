INSERT INTO users (id, name, role) VALUES
  ('usr_001', '销售A', 'sales'),
  ('usr_002', '销售B', 'sales'),
  ('mgr_001', '主管', 'manager')
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  role = VALUES(role);

INSERT INTO customers (
  id, name, source, owner_id, stage, intent, concerns, tags,
  profile_summary, last_contact_at, pending_tasks, risk_flags
) VALUES
  (
    'cus_001',
    '王女士',
    '企业微信',
    'usr_001',
    'price_objection',
    'high',
    JSON_ARRAY('价格', '售后'),
    JSON_ARRAY('价格敏感', '需要案例'),
    '关注价格和售后保障，近期购买意向较强。',
    '2026-05-28 09:30:00',
    1,
    JSON_ARRAY('涉及价格承诺，需人工确认')
  ),
  (
    'cus_002',
    '李先生',
    '上传聊天记录',
    'usr_001',
    'silent',
    'medium',
    JSON_ARRAY('效果', '案例'),
    JSON_ARRAY('需要案例'),
    '已了解产品但 3 天未回复，适合用案例轻触达。',
    '2026-05-25 18:20:00',
    1,
    JSON_ARRAY('客户沉默超过 72 小时')
  ),
  (
    'cus_003',
    '陈总',
    '企业微信',
    'usr_002',
    'high_intent',
    'high',
    JSON_ARRAY('交付周期', '售后'),
    JSON_ARRAY('时间紧急', '关注售后'),
    '明确提出本周内确认方案，主管可关注推进。',
    '2026-05-28 08:50:00',
    1,
    JSON_ARRAY()
  )
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  source = VALUES(source),
  owner_id = VALUES(owner_id),
  stage = VALUES(stage),
  intent = VALUES(intent),
  concerns = VALUES(concerns),
  tags = VALUES(tags),
  profile_summary = VALUES(profile_summary),
  last_contact_at = VALUES(last_contact_at),
  pending_tasks = VALUES(pending_tasks),
  risk_flags = VALUES(risk_flags);

INSERT INTO conversation_messages (id, customer_id, sender_type, sender_name, content, sent_at) VALUES
  ('msg_cus_001_001', 'cus_001', 'customer', '王女士', '这个价格还能优惠吗？', '2026-05-28 09:20:00'),
  ('msg_cus_001_002', 'cus_001', 'sales', '销售A', '我先结合您的需求整理一个更适合的方案。', '2026-05-28 09:22:00'),
  ('msg_cus_002_001', 'cus_002', 'customer', '李先生', '我主要想看看有没有类似案例。', '2026-05-25 18:10:00'),
  ('msg_cus_002_002', 'cus_002', 'sales', '销售A', '可以，我给您整理一个接近场景的案例。', '2026-05-25 18:20:00'),
  ('msg_cus_003_001', 'cus_003', 'customer', '陈总', '如果本周能确认交付边界，我们可以推进。', '2026-05-28 08:45:00'),
  ('msg_cus_003_002', 'cus_003', 'sales', '销售B', '我今天整理交付周期和售后边界给您评估。', '2026-05-28 08:50:00')
ON DUPLICATE KEY UPDATE
  sender_type = VALUES(sender_type),
  sender_name = VALUES(sender_name),
  content = VALUES(content),
  sent_at = VALUES(sent_at);

INSERT INTO followup_tasks (
  id, customer_id, type, status, generated_at, recommendation, feedback
) VALUES
  (
    'task_001',
    'cus_001',
    'price_objection',
    'pending',
    '2026-05-28 10:00:00',
    JSON_OBJECT(
      'customer_stage', 'price_objection',
      'intent_level', 'high',
      'main_concerns', JSON_ARRAY('价格', '效果', '售后'),
      'recommended_action', '解释方案价值并引导预约',
      'script', '您好，刚才您提到比较关注价格，我这边帮您整理了一个更适合您的方案，也可以结合售后保障一起看。',
      'reasoning', '客户连续询问价格和售后，说明有购买兴趣但存在决策顾虑。',
      'risk_flags', JSON_ARRAY('涉及价格承诺，建议人工确认'),
      'next_followup_time', '2026-05-28T16:00:00Z'
    ),
    NULL
  ),
  (
    'task_002',
    'cus_002',
    'silent_reactivation',
    'pending',
    '2026-05-28 10:10:00',
    JSON_OBJECT(
      'customer_stage', 'silent',
      'intent_level', 'medium',
      'main_concerns', JSON_ARRAY('效果', '案例'),
      'recommended_action', '用案例轻触达',
      'script', '您好，前面您提到比较关注实际效果，我整理了一个和您情况接近的案例，您方便的话我发您看一下。',
      'reasoning', '客户已表达兴趣但长时间未回复，直接促单风险较高，适合先提供案例降低压力。',
      'risk_flags', JSON_ARRAY('避免连续催促造成反感'),
      'next_followup_time', '2026-05-28T15:00:00Z'
    ),
    NULL
  ),
  (
    'task_003',
    'cus_003',
    'closing',
    'pending',
    '2026-05-28 10:20:00',
    JSON_OBJECT(
      'customer_stage', 'high_intent',
      'intent_level', 'high',
      'main_concerns', JSON_ARRAY('交付周期', '售后'),
      'recommended_action', '确认决策节点并推动方案评估',
      'script', '陈总，您这边如果希望本周确认，我建议今天先把交付周期和售后边界对齐，我可以整理成一页方案给您内部评估。',
      'reasoning', '客户给出明确时间窗口，需要推进到具体评估动作。',
      'risk_flags', JSON_ARRAY('交付周期不能超出实际能力承诺'),
      'next_followup_time', '2026-05-28T14:00:00Z'
    ),
    NULL
  )
ON DUPLICATE KEY UPDATE
  customer_id = VALUES(customer_id),
  type = VALUES(type),
  status = VALUES(status),
  generated_at = VALUES(generated_at),
  recommendation = VALUES(recommendation),
  feedback = VALUES(feedback);

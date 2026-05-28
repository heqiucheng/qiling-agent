export interface ConversationMessage {
  id: string;
  senderType: "customer" | "sales";
  senderName: string;
  content: string;
  sentAt: string;
}

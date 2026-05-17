import { SidebarLayout } from "@/components/sidebar";

export default function McpLayout({ children }: { children: React.ReactNode }) {
  return <SidebarLayout>{children}</SidebarLayout>;
}

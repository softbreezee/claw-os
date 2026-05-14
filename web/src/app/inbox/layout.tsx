import { SidebarLayout } from "@/components/sidebar";

export default function InboxLayout({ children }: { children: React.ReactNode }) {
  return <SidebarLayout>{children}</SidebarLayout>;
}

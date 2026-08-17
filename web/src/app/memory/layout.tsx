import { SidebarLayout } from "@/components/sidebar";

export default function MemoryLayout({ children }: { children: React.ReactNode }) {
  return <SidebarLayout>{children}</SidebarLayout>;
}

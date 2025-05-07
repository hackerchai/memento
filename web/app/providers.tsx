"use client";

// 使用动态导入避免构建时错误
import dynamic from "next/dynamic";
import { type ReactNode } from "react";

// 动态导入组件
const ThemeProvider = dynamic(() => import("next-themes").then(mod => mod.ThemeProvider), {
  ssr: false 
});

const ToastProvider = dynamic(() => import("@/components/ui/use-toast").then(mod => mod.ToastProvider), {
  ssr: true
});

export function Providers({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <ToastProvider>{children}</ToastProvider>
    </ThemeProvider>
  );
} 
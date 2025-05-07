"use client";

type ToastProps = {
  id?: number;
  title?: string;
  description?: string;
  variant?: "default" | "destructive";
  duration?: number;
};

export function Toaster({ toasts }: { toasts: ToastProps[] }) {
  if (!toasts.length) return null;

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`p-4 rounded-lg shadow-lg flex gap-2 items-start min-w-[260px] max-w-[320px] 
                     animate-in fade-in slide-in-from-top-5 duration-300
                     ${
                      toast.variant === "destructive"
                        ? "bg-red-600 text-white"
                        : "bg-white dark:bg-slate-800 text-gray-900 dark:text-gray-100"
                    }`}
        >
          <div className="flex-1">
            {toast.title && (
              <h3 className="font-semibold text-sm">{toast.title}</h3>
            )}
            {toast.description && (
              <p className="text-xs mt-1">{toast.description}</p>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

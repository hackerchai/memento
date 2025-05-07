"use client";

import { 
  useState, 
  createContext, 
  useContext, 
  type ReactNode,
  type Dispatch,
  type SetStateAction 
} from "react";
import { Toaster } from "./toaster";

type ToastProps = {
  id?: number;
  title?: string;
  description?: string;
  variant?: "default" | "destructive";
  duration?: number;
};

type ToastContextType = {
  toast: (props: ToastProps) => void;
};

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastProps[]>([]);
  const [idCounter, setIdCounter] = useState(0);

  const addToast = (toast: ToastProps) => {
    const uniqueId = Date.now() * 1000 + idCounter;
    setIdCounter(prev => (prev + 1) % 1000);
    
    const newToast = { ...toast, id: uniqueId };
    setToasts((prev: ToastProps[]) => [...prev, newToast]);

    setTimeout(() => {
      setToasts((prev: ToastProps[]) => prev.filter((t: ToastProps) => t.id !== uniqueId));
    }, toast.duration || 3000);
  };

  return (
    <ToastContext.Provider value={{ toast: addToast }}>
      {children}
      <Toaster toasts={toasts} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (context === undefined) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
} 
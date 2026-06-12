import React, { useEffect, useState } from 'react';
import { CheckCircle, XCircle, X } from 'lucide-react';

export type ToastType = 'success' | 'error' | 'info';

interface Toast {
  id: number;
  message: string;
  type: ToastType;
}

let toastCount = 0;
let addToastHandler: (message: string, type: ToastType) => void;

export const showToast = (message: string, type: ToastType = 'success') => {
  if (addToastHandler) addToastHandler(message, type);
};

export const ToastContainer: React.FC = () => {
  const [toasts, setToasts] = useState<Toast[]>([]);

  useEffect(() => {
    addToastHandler = (message, type) => {
      const id = toastCount++;
      setToasts((prev) => [...prev, { id, message, type }]);
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, 5000);
    };
  }, []);

  const removeToast = (id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  return (
    <div className="fixed bottom-6 right-6 z-[9999] flex flex-col gap-3 pointer-events-none">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`pointer-events-auto flex items-center bg-white border rounded-2xl p-4 shadow-2xl min-w-[300px] animate-in slide-in-from-right-10 fade-in duration-300 ${
            t.type === 'success' ? 'border-green-100' : t.type === 'error' ? 'border-red-100' : 'border-blue-100'
          }`}
        >
          <div className="flex-shrink-0 mr-3">
            {t.type === 'success' && <CheckCircle className="w-6 h-6 text-green-500" />}
            {t.type === 'error' && <XCircle className="w-6 h-6 text-red-500" />}
            {t.type === 'info' && <CheckCircle className="w-6 h-6 text-blue-500" />}
          </div>
          <div className="flex-1 text-sm font-bold text-gray-800">{t.message}</div>
          <button onClick={() => removeToast(t.id)} className="ml-4 p-1 hover:bg-gray-50 rounded-lg text-gray-400">
            <X className="w-4 h-4" />
          </button>
        </div>
      ))}
    </div>
  );
};

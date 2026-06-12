import React from 'react';
import { AlertTriangle } from 'lucide-react';

interface ConfirmModalProps {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel: () => void;
  isDanger?: boolean;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm,
  onCancel,
  isDanger = false,
}) => {
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-[100] p-4 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white rounded-2xl w-full max-w-md shadow-2xl overflow-hidden animate-in zoom-in duration-200">
        <div className="p-6 flex flex-col items-center text-center">
          <div className={`w-16 h-16 rounded-full flex items-center justify-center mb-4 ${isDanger ? 'bg-red-50 text-red-500' : 'bg-blue-50 text-blue-500'}`}>
            <AlertTriangle className="w-8 h-8" />
          </div>
          <h3 className="text-xl font-bold text-gray-900 mb-2">{title}</h3>
          <p className="text-gray-500 text-sm leading-relaxed">{message}</p>
        </div>

        <div className="p-4 bg-gray-50 border-t flex gap-3 justify-center">
          <button
            onClick={onCancel}
            className="px-6 py-2.5 bg-white border border-gray-300 rounded-xl text-sm font-bold text-gray-700 hover:bg-gray-100 transition-all shadow-sm"
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            className={`px-6 py-2.5 text-white rounded-xl text-sm font-bold transition-all shadow-md active:scale-95 ${
              isDanger ? 'bg-red-600 hover:bg-red-700 shadow-red-100' : 'bg-blue-600 hover:bg-blue-700 shadow-blue-100'
            }`}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;

import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Loader2, CheckCircle2, CircleDashed, AlertCircle } from 'lucide-react';

interface Task {
  id: string;
  type: string;
  status: string;
  message: string;
  progress: number;
  total: number;
}

const TaskStatus: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isOpen, setIsOpen] = useState(false);

  const fetchTasks = async () => {
    try {
      const res = await axios.get('/api/library/tasks');
      setTasks(res.data || []);
    } catch (err) {
      console.error('Failed to fetch tasks', err);
    }
  };

  useEffect(() => {
    fetchTasks();
    const interval = setInterval(fetchTasks, 3000);
    return () => clearInterval(interval);
  }, []);

  const activeTasks = tasks.filter(t => t.status === 'running' || t.status === 'pending');

  if (tasks.length === 0) return null;

  return (
    <div className="fixed bottom-6 left-6 z-[9999]">
      {/* Mini Indicator */}
      <button 
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-2 px-4 py-2 rounded-full shadow-2xl transition-all border ${
          activeTasks.length > 0 
            ? 'bg-blue-600 border-blue-500 text-white animate-pulse' 
            : 'bg-white border-gray-100 text-gray-600'
        }`}
      >
        {activeTasks.length > 0 ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : (
          <CheckCircle2 className="w-4 h-4 text-green-500" />
        )}
        <span className="text-xs font-black uppercase tracking-widest">
          {activeTasks.length > 0 ? `${activeTasks.length} Active Tasks` : 'Idle'}
        </span>
      </button>

      {/* Task List Dropdown */}
      {isOpen && (
        <div className="absolute bottom-12 left-0 w-80 bg-white rounded-2xl shadow-2xl border border-gray-100 overflow-hidden animate-in slide-in-from-bottom-4 duration-300">
          <div className="p-4 border-b bg-gray-50 flex justify-between items-center">
            <h4 className="text-xs font-black text-gray-500 uppercase tracking-widest">Recent Tasks</h4>
            <span className="text-[10px] text-gray-400 font-bold">{tasks.length} tracked</span>
          </div>
          <div className="max-h-64 overflow-y-auto divide-y divide-gray-50">
            {tasks.slice().reverse().map((t) => (
              <div key={t.id} className="p-3 flex items-start gap-3 hover:bg-gray-50 transition-colors">
                {t.status === 'running' && <Loader2 className="w-4 h-4 text-blue-500 animate-spin mt-0.5" />}
                {t.status === 'pending' && <CircleDashed className="w-4 h-4 text-gray-300 mt-0.5" />}
                {t.status === 'completed' && <CheckCircle2 className="w-4 h-4 text-green-500 mt-0.5" />}
                {t.status === 'failed' && <AlertCircle className="w-4 h-4 text-red-500 mt-0.5" />}
                
                <div className="min-w-0 flex-1">
                  <div className="text-xs font-bold text-gray-800 truncate">{t.type}</div>
                  <div className="text-[10px] text-gray-400 font-medium truncate italic">{t.id}</div>
                  {t.message && <div className="text-[10px] text-gray-500 mt-1">{t.message}</div>}
                  
                  {t.status === 'running' && t.total > 0 && (
                    <div className="mt-2">
                      <div className="flex justify-between items-center mb-1">
                        <span className="text-[9px] font-bold text-blue-500 uppercase tracking-tighter">
                          Progress: {t.progress} / {t.total}
                        </span>
                        <span className="text-[9px] font-black text-blue-600">
                          {Math.round((t.progress / t.total) * 100)}%
                        </span>
                      </div>
                      <div className="w-full bg-gray-100 rounded-full h-1 overflow-hidden">
                        <div 
                          className="bg-blue-500 h-full transition-all duration-500 ease-out rounded-full shadow-[0_0_8px_rgba(59,130,246,0.5)]"
                          style={{ width: `${(t.progress / t.total) * 100}%` }}
                        />
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default TaskStatus;

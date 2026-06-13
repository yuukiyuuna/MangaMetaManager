import React, { useState } from 'react';
import { Plus, Trash2 } from 'lucide-react';
import { api } from '../../api';
import type { LibraryFolder } from '../../api';

interface FolderSectionProps {
  initialFolders: LibraryFolder[];
}

const FolderSection: React.FC<FolderSectionProps> = ({ initialFolders }) => {
  const [folders, setFolders] = useState<LibraryFolder[]>(initialFolders);
  const [newFolderPath, setNewFolderPath] = useState('');

  const refreshFolders = async () => {
    try {
      const res = await api.getLibraryFolders();
      setFolders(res.data);
    } catch (err) {
      console.error(err);
    }
  };

  const addFolder = async () => {
    if (!newFolderPath) return;
    try {
      await api.addLibraryFolder(newFolderPath);
      setNewFolderPath('');
      refreshFolders();
    } catch (err) {
      console.error(err);
    }
  };

  const removeFolder = async (id: number) => {
    try {
      await api.removeLibraryFolder(id);
      refreshFolders();
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <section className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
      <h3 className="text-xl font-semibold text-gray-800 mb-6">Library Folders</h3>
      
      <div className="flex gap-4 mb-6">
        <input
          type="text"
          value={newFolderPath}
          onChange={(e) => setNewFolderPath(e.target.value)}
          className="flex-1 border-gray-300 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 border p-2"
          placeholder="Absolute path to your manga folder"
        />
        <button
          onClick={addFolder}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg flex items-center hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-5 h-5 mr-1" /> Add
        </button>
      </div>

      <div className="space-y-3">
        {folders.map((f) => (
          <div key={f.ID} className="flex justify-between items-center p-3 bg-gray-50 rounded-lg border border-gray-100">
            <span className="text-gray-700 font-mono text-sm">{f.path}</span>
            <button
              onClick={() => removeFolder(f.ID)}
              className="text-red-500 hover:bg-red-50 p-1 rounded transition-colors"
            >
              <Trash2 className="w-5 h-5" />
            </button>
          </div>
        ))}
        {folders.length === 0 && (
          <p className="text-gray-500 text-center py-4">No folders added yet.</p>
        )}
      </div>
    </section>
  );
};

export default FolderSection;

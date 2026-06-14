import React from 'react';
import { Book, Eye, Search, AlertCircle } from 'lucide-react';
import type { MangaBook } from '../../api';

interface BookRowProps {
  book: MangaBook;
  parentTitle: string;
  onViewDetails: () => void;
  onScrape: (contextTitle: string) => void;
}

const BookRow: React.FC<BookRowProps> = ({ book, parentTitle, onViewDetails, onScrape }) => {
  return (
    <div className="flex items-center justify-between p-3 bg-white rounded-xl border border-gray-100 shadow-sm hover:border-blue-200 transition-all group/item">
      <div className="flex items-center flex-1 min-w-0 cursor-pointer" onClick={onViewDetails}>
        <div className="w-8 h-8 bg-blue-50 rounded-lg flex items-center justify-center mr-3 flex-shrink-0">
          <Book className="w-4 h-4 text-blue-400" />
        </div>
        <div className="truncate">
          <div className="flex items-center gap-2">
            <div className="text-sm font-bold text-gray-700 group-hover/item:text-blue-600 transition-colors">{book.filename}</div>
            {book.lastError && (
              <div className="text-red-500" title={book.lastError}>
                <AlertCircle className="w-3.5 h-3.5" />
              </div>
            )}
          </div>
          <div className="text-[10px] text-gray-400 mt-0.5 font-medium">{book.author || 'Unknown Author'} • Vol. {book.volume || '?'}</div>
        </div>
      </div>
      <div className="flex items-center gap-3 ml-4">
        <span className={`text-[10px] font-bold px-2 py-0.5 rounded-md ${book.status === 'Scraped' ? 'text-green-600 bg-green-50' : 'text-gray-400 bg-gray-50'}`}>
          {book.status || 'Pending'}
        </span>
        <div className="flex gap-1.5">
          <button 
            onClick={onViewDetails}
            className="px-2.5 py-1.5 bg-white border border-gray-200 text-gray-600 hover:bg-gray-100 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
          >
            <Eye className="w-3 h-3 mr-1.5" /> DETAILS
          </button>
          <button 
            onClick={() => onScrape(`${parentTitle} ${book.filename}`)}
            className="px-2.5 py-1.5 bg-white border border-blue-100 text-blue-600 hover:bg-blue-600 hover:text-white rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
          >
            <Search className="w-3 h-3 mr-1.5" /> SCRAPE
          </button>
        </div>
      </div>
    </div>
  );
};

export default BookRow;

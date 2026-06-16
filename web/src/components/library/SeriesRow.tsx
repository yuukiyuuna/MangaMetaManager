import React from 'react';
import { ChevronDown, ChevronRight, Eye, Wand2, Search, Trash2, AlertCircle } from 'lucide-react';
import type { MangaSeries } from '../../api';

interface SeriesRowProps {
  series: MangaSeries;
  isExpanded: boolean;
  onToggleExpand: () => void;
  onViewDetails: () => void;
  onAutoScrape: () => void;
  onScrape: () => void;
  onDelete: () => void;
}

const SeriesRow = React.forwardRef<HTMLTableRowElement, SeriesRowProps>(({
  series,
  isExpanded,
  onToggleExpand,
  onViewDetails,
  onAutoScrape,
  onScrape,
  onDelete
}, ref) => {
  return (
    <tr ref={ref} className="hover:bg-blue-50/30 transition-colors group">
      <td className="px-6 py-5 text-center">
        <button onClick={onToggleExpand} className="text-gray-300 hover:text-blue-600 transition-colors">
          {isExpanded ? <ChevronDown className="w-5 h-5" /> : <ChevronRight className="w-5 h-5" />}
        </button>
      </td>
      <td className="px-6 py-5 cursor-pointer" onClick={onViewDetails}>
        <div className="flex items-center gap-2">
          <div className="font-bold text-gray-900 group-hover:text-blue-700 transition-colors">{series.title}</div>
          {series.lastError && (
            <div className="text-red-500" title={series.lastError}>
              <AlertCircle className="w-4 h-4" />
            </div>
          )}
        </div>
        <div className="text-[10px] text-gray-400 font-mono truncate max-w-[200px] mt-0.5">{series.path}</div>
      </td>
      <td className="px-6 py-5 text-sm text-gray-600 font-medium">
        {series.author || <span className="text-gray-300 italic">Unknown</span>}
      </td>
      <td className="px-6 py-5 text-sm text-gray-500 font-bold">{series.books?.length || 0}</td>
      <td className="px-6 py-5">
        <span className={`px-2.5 py-1 text-[10px] font-black uppercase tracking-wider rounded-lg shadow-sm ${
          series.status === 'Scraped' ? 'bg-green-100 text-green-700 border border-green-200' : 'bg-yellow-50 text-yellow-700 border border-yellow-100'
        }`}>
          {series.status || 'Unscraped'}
        </span>
      </td>
      <td className="px-6 py-5">
        <div className="flex justify-end gap-2">
          <button 
            onClick={onViewDetails}
            className="px-3 py-1.5 bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
            title="View Details"
          >
            <Eye className="w-3.5 h-3.5 mr-1.5" /> DETAILS
          </button>
          <button 
            onClick={onAutoScrape}
            className="px-3 py-1.5 bg-white border border-purple-200 text-purple-600 hover:bg-purple-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
            title="Auto Scrape All Books"
          >
            <Wand2 className="w-3.5 h-3.5 mr-1.5" /> AUTO
          </button>
          <button 
            onClick={onScrape}
            className="px-3 py-1.5 bg-white border border-blue-200 text-blue-600 hover:bg-blue-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
          >
            <Search className="w-3.5 h-3.5 mr-1.5" /> SCRAPE
          </button>
          <button 
            onClick={onDelete}
            className="px-3 py-1.5 bg-white border border-red-200 text-red-500 hover:bg-red-50 rounded-lg text-[10px] font-black transition-all shadow-sm flex items-center"
            title="Remove from Database"
          >
            <Trash2 className="w-3.5 h-3.5 mr-1.5" /> REMOVE
          </button>
        </div>
      </td>
    </tr>
  );
});

SeriesRow.displayName = 'SeriesRow';

export default SeriesRow;

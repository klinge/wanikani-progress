import React from 'react';
import { Egg, Flower2, Flower, Sparkles, Flame } from 'lucide-react';

const stages = [
  { name: 'Apprentice',   icon: Egg,      radical: 'bg-blue-500',   kanji: 'bg-pink-500',    vocab: 'bg-purple-500', total: 'bg-gray-500' },
  { name: 'Guru',         icon: Flower2,  radical: 'bg-blue-500',   kanji: 'bg-pink-500',    vocab: 'bg-purple-500', total: 'bg-gray-500' },
  { name: 'Master',       icon: Flower,   radical: 'bg-blue-500',   kanji: 'bg-pink-500',    vocab: 'bg-purple-500', total: 'bg-gray-500' },
  { name: 'Enlightened',  icon: Sparkles, radical: 'bg-blue-500',   kanji: 'bg-pink-500',    vocab: 'bg-purple-500', total: 'bg-gray-500' },
  { name: 'Burned',       icon: Flame,    radical: 'bg-blue-500',   kanji: 'bg-pink-500',    vocab: 'bg-purple-500', total: 'bg-gray-500' },
];

export default function ItemSpreadCard({ data }) {
  return (
    <div className="w-full max-w-4xl mx-auto">
      <div className="bg-white rounded-2xl shadow-lg overflow-hidden">
        {/* Header */}
        <div className="px-8 pt-8 pb-6">
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-bold text-gray-900">Item Spread</h2>
            <div className="flex items-center gap-6 text-sm">
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-blue-500 rounded"></div>
                <span className="text-gray-600">Radicals</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-pink-500 rounded"></div>
                <span className="text-gray-600">Kanji</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-4 h-4 bg-purple-500 rounded"></div>
                <span className="text-gray-600">Vocabulary</span>
              </div>
            </div>
          </div>
        </div>

        {/* Rows */}
        <div className="px-8 pb-8 space-y-3">
          {stages.map((stage) => {
            const items = {
              apprentice:   data.apprentice,
              guru:         data.guru,
              master:       data.master,
              enlightened:  data.enlightened,
              burned:       data.burned,
            }[stage.name.toLowerCase()];

            const Icon = stage.icon;

            return (
              <div
                key={stage.name}
                className="flex items-center bg-gray-100 rounded-xl px-2 py-2 shadow-sm border border-gray-300/70"
              >
                {/* Stage name + icon */}
                <div className="flex items-center gap-4 w-48">
                  <Icon className="w-8 h-6 text-gray-600" />
                  <span className="text-lg font-medium text-gray-800">{stage.name}</span>
                </div>

                {/* Numbers */}
                <div className="flex-1 flex items-center gap-3 justify-end max-w-lg">
                  <div className={`w-16 h-8 flex items-center justify-center rounded-lg text-white font-bold text-lg ${stage.radical}`}>
                    {items.radical}
                  </div>
                  <div className={`w-16 h-8 flex items-center justify-center rounded-lg text-white font-bold text-lg ${stage.kanji}`}>
                    {items.kanji}
                  </div>
                  <div className={`w-16 h-8 flex items-center justify-center rounded-lg text-white font-bold text-lg ${stage.vocab}`}>
                    {items.vocabulary}
                  </div>
                </div>

                {/* Total */}
                <div className={`ml-8 w-20 h-8 flex items-center justify-center rounded-lg text-white font-bold text-xl ${stage.total}`}>
                  {items.total}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
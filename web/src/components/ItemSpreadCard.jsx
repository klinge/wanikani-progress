import React, { useState, useEffect } from 'react';
import { Egg, Turtle, Flower, Sparkles, Flame } from 'lucide-react';
import { yesterdayISO } from '../utils/date';
import { wanikaniAPI } from '../services/api';

const stages = [
    { name: 'Apprentice', icon: Egg, radical: 'bg-blue-500', kanji: 'bg-pink-500', vocab: 'bg-purple-500', total: 'bg-gray-500' },
    { name: 'Guru', icon: Turtle, radical: 'bg-blue-500', kanji: 'bg-pink-500', vocab: 'bg-purple-500', total: 'bg-gray-500' },
    { name: 'Master', icon: Flower, radical: 'bg-blue-500', kanji: 'bg-pink-500', vocab: 'bg-purple-500', total: 'bg-gray-500' },
    { name: 'Enlightened', icon: Sparkles, radical: 'bg-blue-500', kanji: 'bg-pink-500', vocab: 'bg-purple-500', total: 'bg-gray-500' },
    { name: 'Burned', icon: Flame, radical: 'bg-blue-500', kanji: 'bg-pink-500', vocab: 'bg-purple-500', total: 'bg-gray-500' },
];

export default function ItemSpreadCard() {

    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // Fetch data from api on mount
    useEffect(() => {
        const fetchTodaySnapshot = async () => {
            try {
                setLoading(true);

                const yesterday = yesterdayISO(); // "2025-12-04"

                const response = await wanikaniAPI.getAssignmentSnapshots(yesterday, yesterday);
                // response.data is an object like: { "2025-12-04": { apprentice: {...}, ... } }
                // console.info('Fetched snapshots:', response.data);
                const yesterdayData = response.data[yesterday];

                if (!yesterdayData) {
                    throw new Error('No snapshot found for today');
                }

                setData(yesterdayData);
            } catch (err) {
                console.error(err);
                setError('Could not load today\'s item spread');
            } finally {
                setLoading(false);
            }
        };

        fetchTodaySnapshot();
    }, []);

    if (loading) {
        return (
            <div className="w-full max-w-4xl mx-auto p-8">
                <div className="bg-white rounded-2xl shadow-lg p-8">
                    <div className="animate-pulse">
                        <div className="h-8 bg-gray-200 rounded w-48 mb-8"></div>
                        <div className="space-y-4">
                            {[1, 2, 3, 4, 5].map(i => (
                                <div key={i} className="h-20 bg-gray-100 rounded-xl"></div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="w-full max-w-4xl mx-auto p-8">
                <div className="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
                    <p className="text-red-800">{error}</p>
                </div>
            </div>
        );
    }

    // Rendering the data
    return (
        <div className="w-full h-full">
            <div className="bg-white rounded-2xl shadow-lg overflow-hidden h-full flex flex-col">
                {/* Header */}
                <div className="px-6 py-2">
                    <div className="flex items-center justify-between">
                        <h2 className="text-2xl font-bold text-gray-900">Item Spread</h2>
                        <div className="flex items-center gap-4 text-sm">
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
                <div className="px-6 pb-6 space-y-2 flex-1 overflow-y-auto">
                    {stages.map((stage) => {
                        const items = {
                            apprentice: data.apprentice,
                            guru: data.guru,
                            master: data.master,
                            enlightened: data.enlightened,
                            burned: data.burned,
                        }[stage.name.toLowerCase()];

                        const Icon = stage.icon;

                        return (
                            <div
                                key={stage.name}
                                className="flex items-center bg-gray-200 rounded-lg px-2 py-2 shadow-sm border border-gray-300/70"
                            >
                                {/* Stage name + icon */}
                                <div className="flex items-center gap-3 w-36">
                                    <Icon className="w-6 h-5 text-gray-600" />
                                    <span className="text-lg text-gray-800">{stage.name}</span>
                                </div>

                                {/* Numbers */}
                                <div className="flex-1 flex items-center gap-1.5 justify-end">
                                    <div className={`w-10 h-6 flex items-center justify-center rounded text-white font-bold text-sm ${stage.radical}`}>
                                        {items.radical}
                                    </div>
                                    <div className={`w-10 h-6 flex items-center justify-center rounded text-white font-bold text-sm ${stage.kanji}`}>
                                        {items.kanji}
                                    </div>
                                    <div className={`w-10 h-6 flex items-center justify-center rounded text-white font-bold text-sm ${stage.vocab}`}>
                                        {items.vocabulary}
                                    </div>
                                </div>

                                {/* Total */}
                                <div className={`ml-3 w-16 h-6 flex items-center justify-center rounded text-white font-bold text-sm ${stage.total}`}>
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
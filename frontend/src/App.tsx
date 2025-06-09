import { useState, useEffect } from 'react';
import { GetSlides, LoadPresentation, SendMessageToAI } from "../wailsjs/go/main/App";
import ChatPanel from './components/ChatPanel';
import SlideViewer from './components/SlideViewer';

function App() {
    const [slides, setSlides] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);
    const [chatOpen, setChatOpen] = useState(false);

    useEffect(() => {
        // Load initial slides if they exist
        loadSlides();
    }, []);

    const loadSlides = async () => {
        try {
            const slideList = await GetSlides();
            setSlides(slideList);
        } catch (error) {
            console.error('Failed to load slides:', error);
        }
    };

    const handleLoadPresentation = async () => {
        setLoading(true);
        try {
            // For now, load the sample presentation
            const slideList = await LoadPresentation('original_ppt.pptx');
            setSlides(slideList);
        } catch (error) {
            console.error('Failed to load presentation:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleSendMessage = async (message: string) => {
        try {
            const response = await SendMessageToAI(message);
            // Reload slides after AI interaction in case they were modified
            await loadSlides();
            return response;
        } catch (error) {
            console.error('Failed to send message to AI:', error);
            throw error;
        }
    };

    return (
        <div className="flex h-screen bg-gray-900 text-white">
            {/* Main content area */}
            <div className="flex-1 flex flex-col">
                {/* Header */}
                <div className="bg-gray-800 p-4 border-b border-gray-700">
                    <div className="flex items-center justify-between">
                        <h1 className="text-xl font-bold">SlidePilot</h1>
                        <div className="flex space-x-2">
                            <button
                                onClick={handleLoadPresentation}
                                disabled={loading}
                                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-sm font-medium disabled:opacity-50"
                            >
                                {loading ? 'Loading...' : 'Load Sample Presentation'}
                            </button>
                            <button
                                onClick={() => setChatOpen(!chatOpen)}
                                className="px-4 py-2 bg-green-600 hover:bg-green-700 rounded text-sm font-medium"
                            >
                                {chatOpen ? 'Close Chat' : 'Open AI Chat'}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Content */}
                <div className="flex-1 p-4">
                    {slides.length === 0 ? (
                        <div className="flex items-center justify-center h-full">
                            <div className="text-center">
                                <h2 className="text-2xl font-semibold mb-4 text-gray-400">No slides loaded</h2>
                                <p className="text-gray-500 mb-6">Load a presentation to get started</p>
                                <button
                                    onClick={handleLoadPresentation}
                                    disabled={loading}
                                    className="px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-medium disabled:opacity-50"
                                >
                                    {loading ? 'Loading...' : 'Load Sample Presentation'}
                                </button>
                            </div>
                        </div>
                    ) : (
                        <SlideViewer slides={slides} onRefresh={loadSlides} />
                    )}
                </div>
            </div>

            {/* Chat panel */}
            {chatOpen && (
                <div className="w-96 border-l border-gray-700">
                    <ChatPanel onSendMessage={handleSendMessage} />
                </div>
            )}
        </div>
    );
}

export default App;

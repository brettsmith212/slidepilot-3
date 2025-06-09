import { useState } from 'react';

interface SlideViewerProps {
    slides: string[];
    onRefresh: () => void;
}

const SlideViewer: React.FC<SlideViewerProps> = ({ slides, onRefresh }) => {
    const [currentSlide, setCurrentSlide] = useState(0);

    const nextSlide = () => {
        setCurrentSlide((prev) => (prev + 1) % slides.length);
    };

    const prevSlide = () => {
        setCurrentSlide((prev) => (prev - 1 + slides.length) % slides.length);
    };

    const goToSlide = (index: number) => {
        setCurrentSlide(index);
    };

    if (slides.length === 0) {
        return null;
    }

    return (
        <div className="flex h-full">
            {/* Slide thumbnails */}
            <div className="w-48 bg-gray-800 overflow-y-auto p-2 border-r border-gray-700">
                <div className="flex flex-col space-y-2">
                    {slides.map((slide, index) => (
                        <button
                            key={index}
                            onClick={() => goToSlide(index)}
                            className={`relative border-2 rounded transition-colors ${
                                currentSlide === index
                                    ? 'border-blue-500 bg-blue-900/20'
                                    : 'border-gray-600 hover:border-gray-500'
                            }`}
                        >
                            <img
                                src={`slidepilot-3://${slide}`}
                                alt={`Slide ${index + 1}`}
                                className="w-full aspect-[4/3] object-cover rounded"
                            />
                            <div className="absolute bottom-1 left-1 bg-black bg-opacity-70 text-white text-xs px-1 py-0.5 rounded">
                                {index + 1}
                            </div>
                        </button>
                    ))}
                </div>
                <button
                    onClick={onRefresh}
                    className="w-full mt-4 px-3 py-2 bg-gray-700 hover:bg-gray-600 rounded text-sm"
                >
                    Refresh Slides
                </button>
            </div>

            {/* Main slide view */}
            <div className="flex-1 flex flex-col">
                <div className="flex-1 flex items-center justify-center bg-gray-850 p-4">
                    <div className="max-w-4xl max-h-full">
                        <img
                            src={`slidepilot-3://${slides[currentSlide]}`}
                            alt={`Slide ${currentSlide + 1}`}
                            className="max-w-full max-h-full object-contain rounded-lg shadow-lg"
                        />
                    </div>
                </div>

                {/* Navigation controls */}
                <div className="bg-gray-800 p-4 border-t border-gray-700">
                    <div className="flex items-center justify-between">
                        <button
                            onClick={prevSlide}
                            disabled={slides.length <= 1}
                            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Previous
                        </button>
                        
                        <div className="flex items-center space-x-2">
                            <span className="text-sm text-gray-300">
                                Slide {currentSlide + 1} of {slides.length}
                            </span>
                        </div>
                        
                        <button
                            onClick={nextSlide}
                            disabled={slides.length <= 1}
                            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Next
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default SlideViewer;

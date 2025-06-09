import { useState, useEffect } from "react";
import {
  GetSlides,
  LoadPresentation,
  SendMessageToAI,
  GetSlideImageQuiet,
  GetSlideImageAsBase64,
} from "../wailsjs/go/main/App";
import ChatPanel from "./components/ChatPanel";

function App() {
  const [slides, setSlides] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [chatOpen, setChatOpen] = useState(false);
  const [currentSlide, setCurrentSlide] = useState(0);
  const [currentSlideImage, setCurrentSlideImage] = useState<string>("");

  useEffect(() => {
    // Load initial slides if they exist
    loadSlides();
  }, []);

  useEffect(() => {
    // Load current slide image when slide changes
    if (slides.length > 0 && slides[currentSlide]) {
      loadCurrentSlideImage();
    }
  }, [currentSlide, slides]);

  const loadCurrentSlideImage = async () => {
    if (slides.length === 0) return;

    try {
      // First call the quiet method to cache data without logging base64
      const status = await GetSlideImageQuiet(slides[currentSlide]);

      // Then get the actual base64 data from cache
      if (
        status === "CACHED_BASE64_DATA_AVAILABLE" ||
        status === "BASE64_DATA_LOADED"
      ) {
        const imageData = await GetSlideImageAsBase64(slides[currentSlide]);
        setCurrentSlideImage(imageData);
      }
    } catch (error) {
      console.error("Failed to load slide image:", error);
      setCurrentSlideImage("");
    }
  };

  const loadSlides = async () => {
    try {
      const slideList = await GetSlides();
      setSlides(slideList);
      // Reset current slide image when slides change
      if (slideList.length > 0) {
        setCurrentSlideImage("");
      }
    } catch (error) {
      console.error("Failed to load slides:", error);
    }
  };

  const handleLoadPresentation = async () => {
    setLoading(true);
    try {
      // For now, load the sample presentation
      const slideList = await LoadPresentation("original_ppt.pptx");
      setSlides(slideList);
      setCurrentSlide(0);
    } catch (error) {
      console.error("Failed to load presentation:", error);
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
      console.error("Failed to send message to AI:", error);
      throw error;
    }
  };

  const nextSlide = () => {
    if (slides.length > 0) {
      setCurrentSlide((prev) => (prev + 1) % slides.length);
    }
  };

  const prevSlide = () => {
    if (slides.length > 0) {
      setCurrentSlide((prev) => (prev - 1 + slides.length) % slides.length);
    }
  };

  return (
    <div className="flex h-screen bg-white">
      {/* Main content area */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div className="flex items-center space-x-3">
            <h1 className="text-xl font-bold text-gray-900">SlidePilot</h1>
            <span className="text-gray-500">AI-Powered PowerPoint Editor</span>
          </div>
          <button
            onClick={() => setChatOpen(!chatOpen)}
            className="p-2 hover:bg-gray-100 rounded-md transition-colors"
          >
            {chatOpen ? (
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            ) : (
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
            )}
          </button>
        </div>

        {/* Toolbar */}
        <div className="flex items-center justify-between px-6 py-3 border-b border-gray-200 bg-gray-50">
          <div className="flex items-center space-x-3">
            <button
              onClick={handleLoadPresentation}
              disabled={loading}
              className="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <svg
                className="w-4 h-4 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                />
              </svg>
              {loading ? "Loading..." : "Open Presentation"}
            </button>

            <div className="flex items-center space-x-2 ml-6">
              <button className="p-2 hover:bg-gray-200 rounded-md transition-colors">
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                  />
                </svg>
              </button>
              <button className="px-3 py-1 hover:bg-gray-200 rounded-md transition-colors text-sm font-medium">
                Fit
              </button>
              <button className="p-2 hover:bg-gray-200 rounded-md transition-colors">
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                  />
                </svg>
              </button>
              <button className="p-2 hover:bg-gray-200 rounded-md transition-colors">
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 8V4m0 0h4m-4 0l5.657 5.657M20 8V4m0 0h-4m4 0l-5.657 5.657M4 16v4m0 0h4m-4 0l5.657-5.657M20 16v4m0 0h-4m4 0l-5.657-5.657"
                  />
                </svg>
              </button>
            </div>
          </div>

          {slides.length > 0 && (
            <div className="text-sm font-medium text-gray-700">
              Slide {currentSlide + 1} of {slides.length}
            </div>
          )}
        </div>

        {/* Main Content */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {slides.length === 0 ? (
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center max-w-lg">
                <h2 className="text-3xl font-semibold text-gray-900 mb-4">
                  Welcome to SlidePilot
                </h2>
                <p className="text-lg text-gray-600 mb-8">
                  Your AI-powered presentation editor
                </p>
                <p className="text-gray-500 mb-8">
                  Click 'Open Presentation' to load a PowerPoint file and see
                  the slide parsing in action.
                </p>
              </div>
            </div>
          ) : (
            <div className="flex-1 flex items-center justify-center p-4 bg-gray-50 min-h-0">
              <div className="max-w-4xl w-full h-full bg-white rounded-lg shadow-sm border border-gray-200 p-4 flex items-center justify-center">
                {currentSlideImage ? (
                  <img
                    src={currentSlideImage}
                    alt={`Slide ${currentSlide + 1}`}
                    className="max-w-full max-h-full object-contain"
                  />
                ) : (
                  <div className="w-full h-32 flex items-center justify-center text-gray-500">
                    Loading slide...
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Bottom Navigation */}
        {slides.length > 0 && (
          <div className="border-t border-gray-200 bg-white p-4">
            <div className="flex items-center justify-between mb-4">
              <button
                onClick={prevSlide}
                className="inline-flex items-center px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
              >
                <svg
                  className="w-4 h-4 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M15 19l-7-7 7-7"
                  />
                </svg>
                Previous
              </button>

              <div className="text-sm font-medium text-gray-700">
                {currentSlide + 1} of {slides.length}
              </div>

              <button
                onClick={nextSlide}
                className="inline-flex items-center px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-md transition-colors"
              >
                Next
                <svg
                  className="w-4 h-4 ml-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </button>
            </div>

            {/* Slide Thumbnails */}
            <div className="flex justify-center space-x-2">
              {slides.map((_, index) => (
                <button
                  key={index}
                  onClick={() => setCurrentSlide(index)}
                  className={`w-12 h-9 border-2 rounded text-xs font-medium transition-colors ${
                    currentSlide === index
                      ? "border-blue-500 bg-blue-50 text-blue-700"
                      : "border-gray-300 bg-white text-gray-700 hover:border-gray-400"
                  }`}
                >
                  {index + 1}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* AI Chat Panel */}
      {chatOpen && (
        <div className="w-96 border-l border-gray-200 bg-white">
          <ChatPanel onSendMessage={handleSendMessage} />
        </div>
      )}
    </div>
  );
}

export default App;

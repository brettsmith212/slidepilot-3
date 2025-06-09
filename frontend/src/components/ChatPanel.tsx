import { useState, useRef, useEffect } from 'react';

interface ChatMessage {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    timestamp: Date;
}

interface ChatPanelProps {
    onSendMessage: (message: string) => Promise<string>;
}

const ChatPanel: React.FC<ChatPanelProps> = ({ onSendMessage }) => {
    const [messages, setMessages] = useState<ChatMessage[]>([
        {
            id: '1',
            role: 'assistant',
            content: 'Hello! I\'m your AI presentation assistant. I can help you edit slides, improve content, and navigate your presentation.',
            timestamp: new Date()
        },
        {
            id: '2',
            role: 'assistant',
            content: 'Try asking me to:\n• "Edit the title of this slide"\n• "Go to slide 3"\n• "Analyze this presentation"\n• "Make the text larger"',
            timestamp: new Date()
        }
    ]);
    const [inputMessage, setInputMessage] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    useEffect(() => {
        scrollToBottom();
    }, [messages]);

    const handleSendMessage = async () => {
        if (!inputMessage.trim() || isLoading) return;

        const userMessage: ChatMessage = {
            id: Date.now().toString(),
            role: 'user',
            content: inputMessage.trim(),
            timestamp: new Date()
        };

        setMessages(prev => [...prev, userMessage]);
        setInputMessage('');
        setIsLoading(true);

        try {
            const response = await onSendMessage(inputMessage.trim());
            
            const assistantMessage: ChatMessage = {
                id: (Date.now() + 1).toString(),
                role: 'assistant',
                content: response,
                timestamp: new Date()
            };

            setMessages(prev => [...prev, assistantMessage]);
        } catch (error) {
            const errorMessage: ChatMessage = {
                id: (Date.now() + 1).toString(),
                role: 'assistant',
                content: 'Sorry, I encountered an error while processing your request. Please try again.',
                timestamp: new Date()
            };

            setMessages(prev => [...prev, errorMessage]);
            console.error('Chat error:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSendMessage();
        }
    };

    const formatTimestamp = (timestamp: Date) => {
        return timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    };

    const handleSuggestionClick = (suggestion: string) => {
        setInputMessage(suggestion);
    };

    const suggestions = [
        "Edit the title",
        "Go to slide 2", 
        "Make the text larger"
    ];

    return (
        <div className="flex flex-col h-full bg-white">
            {/* Header */}
            <div className="p-4 border-b border-gray-200">
                <div className="flex items-center space-x-2 mb-1">
                    <h2 className="text-lg font-semibold text-gray-900">AI Assistant</h2>
                    <div className="flex items-center space-x-1">
                        <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                        <span className="text-sm text-gray-600">Online</span>
                    </div>
                </div>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {messages.map((message) => (
                    <div key={message.id} className="flex items-start space-x-3">
                        {message.role === 'assistant' && (
                            <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-sm font-medium">
                                AI
                            </div>
                        )}
                        <div className={`flex-1 ${message.role === 'user' ? 'ml-8' : ''}`}>
                            <div
                                className={`rounded-lg p-3 ${
                                    message.role === 'user'
                                        ? 'bg-blue-600 text-white ml-8'
                                        : 'bg-gray-100 text-gray-900'
                                }`}
                            >
                                <div className="text-sm whitespace-pre-wrap">{message.content}</div>
                            </div>
                            <div className="text-xs text-gray-500 mt-1">
                                {formatTimestamp(message.timestamp)}
                            </div>
                        </div>
                    </div>
                ))}
                
                {isLoading && (
                    <div className="flex items-start space-x-3">
                        <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-sm font-medium">
                            AI
                        </div>
                        <div className="bg-gray-100 rounded-lg p-3">
                            <div className="flex items-center space-x-2">
                                <div className="flex space-x-1">
                                    <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                                    <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                                    <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                                </div>
                                <span className="text-sm text-gray-600">AI is thinking...</span>
                            </div>
                        </div>
                    </div>
                )}
                
                <div ref={messagesEndRef} />
            </div>

            {/* Input */}
            <div className="p-4 border-t border-gray-200">
                <div className="flex space-x-2 mb-3">
                    <input
                        value={inputMessage}
                        onChange={(e) => setInputMessage(e.target.value)}
                        onKeyPress={handleKeyPress}
                        placeholder="Ask me to edit slides, navigate..."
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
                        disabled={isLoading}
                    />
                    <button
                        onClick={handleSendMessage}
                        disabled={!inputMessage.trim() || isLoading}
                        className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed rounded-lg text-white font-medium transition-colors"
                    >
                        Send
                    </button>
                </div>
                
                {/* Suggestions */}
                <div className="text-xs text-gray-500 mb-2">
                    Try: {suggestions.map((suggestion, index) => (
                        <span key={index}>
                            <button
                                onClick={() => handleSuggestionClick(suggestion)}
                                className="text-blue-600 hover:underline"
                            >
                                "{suggestion}"
                            </button>
                            {index < suggestions.length - 1 && ', '}
                        </span>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default ChatPanel;

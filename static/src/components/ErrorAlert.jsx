export default function ErrorAlert({ message, onClose }) {
    return (
        <div className="fixed top-4 left-4 right-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
            <div className="flex justify-between items-center">
                <span>{message}</span>
                <button
                    onClick={onClose}
                    className="text-red-500 hover:text-red-700"
                >
                    ×
                </button>
            </div>
        </div>
    )
}
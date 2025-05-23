import React, { createContext, useReducer, useContext, ReactNode, useCallback } from 'react';

// 1. Define State and Action Types
interface Notification {
  id: number;
  message: string;
  type: 'success' | 'error' | 'info'; // Add more types as needed
  duration?: number; // Optional: duration in ms, auto-hides if set
}

interface NotificationState {
  notifications: Notification[];
}

type NotificationAction =
  | { type: 'ADD_NOTIFICATION'; payload: Omit<Notification, 'id'> }
  | { type: 'REMOVE_NOTIFICATION'; payload: { id: number } }
  | { type: 'CLEAR_NOTIFICATIONS' };

// 2. Define Initial State
const initialNotificationState: NotificationState = {
  notifications: [],
};

// 3. Create Context
interface NotificationContextProps {
  notificationState: NotificationState;
  dispatchNotificationAction: React.Dispatch<NotificationAction>;
  addNotification: (message: string, type: Notification['type'], duration?: number) => void; // Convenience function
}

const NotificationContext = createContext<NotificationContextProps | undefined>(undefined);

let notificationIdCounter = 0;

// 4. Implement Reducer
const notificationReducer = (state: NotificationState, action: NotificationAction): NotificationState => {
  switch (action.type) {
    case 'ADD_NOTIFICATION':
      return {
        ...state,
        notifications: [...state.notifications, { ...action.payload, id: notificationIdCounter++ }],
      };
    case 'REMOVE_NOTIFICATION':
      return {
        ...state,
        notifications: state.notifications.filter(notification => notification.id !== action.payload.id),
      };
    case 'CLEAR_NOTIFICATIONS':
      return {
        ...state,
        notifications: [],
      };
    default:
      return state;
  }
};

// 5. Implement Context Provider
interface NotificationProviderProps {
  children: ReactNode;
}

export const NotificationProvider: React.FC<NotificationProviderProps> = ({ children }) => {
  const [notificationState, dispatchNotificationAction] = useReducer(notificationReducer, initialNotificationState);

  const addNotification = useCallback((message: string, type: Notification['type'], duration?: number) => {
    const newNotification = { message, type, duration };
    dispatchNotificationAction({ type: 'ADD_NOTIFICATION', payload: newNotification });

    if (duration) {
      const currentId = notificationIdCounter -1; // get current id
      setTimeout(() => {
        dispatchNotificationAction({ type: 'REMOVE_NOTIFICATION', payload: { id: currentId } });
      }, duration);
    }
  }, []);


  return (
    <NotificationContext.Provider value={{ notificationState, dispatchNotificationAction, addNotification }}>
      {children}
    </NotificationContext.Provider>
  );
};

// 6. Create Custom Hook
export const useNotification = (): NotificationContextProps => {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotification must be used within a NotificationProvider');
  }
  return context;
};

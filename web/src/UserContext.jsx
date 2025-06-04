import { createContext, useContext } from "react";
import useFetchUser from "./hooks/useFetchUser";

const UserContext = createContext();

export const useUser = () => {
  return useContext(UserContext);
};

// eslint-disable-next-line react/prop-types
export const UserProvider = ({ children }) => {
  const { user, loading } = useFetchUser();

  console.log("This is from Provider: ", user);

  if (loading) {
    return <div>Loading...</div>; // Or any fallback UI
  }

  return (
    <UserContext.Provider value={{ user }}>{children}</UserContext.Provider>
  );
};

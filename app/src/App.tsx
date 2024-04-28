import React from "react";
import {Route, Routes} from "react-router";
import {BrowserRouter} from "react-router-dom";
import {LoginPage} from "./pages/LoginPage";

export function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/login" element={<LoginPage/>}/>
            </Routes>
        </BrowserRouter>
    )
}

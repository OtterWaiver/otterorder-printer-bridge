import { useEffect, useState } from 'react';
import './App.css';
import { GetServerStatus, GetPreferences, SavePreferences } from "../wailsjs/go/main/App";

function App() {
    const [statusText, setStatusText] = useState("Please enter your name below 👇");
    const [ip, setIp] = useState('');
    const [port, setPort] = useState('');

    const updateIp = (e) => setIp(e.target.value);
    const updatePort = (e) => setPort(e.target.value);
    const updateStatusText = (result) => setStatusText(result);
    const [isSaving, setIsSaving] = useState(false);
    const [hasSaved, setHasSaved] = useState(false);

    useEffect(() => {
        GetServerStatus().then(updateStatusText);
        GetPreferences().then((prefs) => {
            setIp(prefs.printer_ip);
            setPort(prefs.printer_port);
        });
    }, [])

    const save = () => {
        setIsSaving(true);
        SavePreferences({ printer_ip: ip, printer_port: port }).then(() => {
            GetServerStatus().then(updateStatusText);
            setHasSaved(true);
            setTimeout(() => setHasSaved(false), 2000);
        }).catch(console.error).finally(() => {
            setIsSaving(false);
        })
    }

    return (
        <div id="App">
            <div>
                <div id="result" className="result">{statusText}</div>
                <div id="input" className="input-box">
                    <input value={ip} placeholder='Enter printer IP' id="ip" className="input" onChange={updateIp} autoComplete="off" name="input" type="text" />
                    <input value={port} placeholder='Enter printer port' id="port" className="input" onChange={updatePort} autoComplete="off" name="input" type="text" />
                    <button className="btn" onClick={() => save()}>{
                        isSaving ? "Saving..." : hasSaved ? "Saved!" : "Save"
                    }</button>
                </div>
            </div>
        </div>
    )
}

export default App

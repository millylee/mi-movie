import {useState, useEffect} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Greet, GetHomePage, GetWebViewConfig, GetConfigInfo} from "../wailsjs/go/main/App";

function App() {
    const [resultText, setResultText] = useState("Please enter your name below 👇");
    const [name, setName] = useState('');
    const [homePage, setHomePage] = useState('');
    const [config, setConfig] = useState<any>({});
    const [showWebView, setShowWebView] = useState(false);
    const updateName = (e: any) => setName(e.target.value);
    const updateResultText = (result: string) => setResultText(result);

    function greet() {
        Greet(name).then(updateResultText);
    }

    function navigateToHomePage() {
        if (homePage && homePage !== '') {
            setShowWebView(true);
        }
    }

    function goBack() {
        setShowWebView(false);
    }

    useEffect(() => {
        // Load configuration
        GetConfigInfo().then(config => {
            setConfig(config);
        });
        
        GetHomePage().then(page => {
            setHomePage(page);
            if (page && page !== '' && page !== "https://www.iyf.tv") {
                setTimeout(() => {
                    setShowWebView(true);
                }, 1000);
            }
        });
    }, []);

    // Show webview if custom home page is set
    if (showWebView && homePage) {
        return (
            <div style={{width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column'}}>
                <div style={{padding: '10px', backgroundColor: '#f0f0f0', borderBottom: '1px solid #ccc'}}>
                    <button onClick={goBack} style={{marginRight: '10px'}}>返回</button>
                    <span>当前页面: {homePage}</span>
                    <div style={{fontSize: '12px', color: '#666', marginTop: '5px'}}>
                        代理: {config.proxyServer || '未设置'} | 数据目录: {config.userData || '默认位置'}
                    </div>
                </div>
                <iframe 
                    src={homePage} 
                    style={{width: '100%', height: '100%', border: 'none'}}
                    title="Web Content"
                />
            </div>
        );
    }

    return (
        <div id="App">
            <img src={logo} id="logo" alt="logo"/>
            <div id="result" className="result">{resultText}</div>
            <div id="input" className="input-box">
                <input id="name" className="input" onChange={updateName} autoComplete="off" name="input" type="text"/>
                <button className="btn" onClick={greet}>Greet</button>
            </div>
            <div className="navigation">
                <button className="btn" onClick={navigateToHomePage}>
                    Go to Home Page {homePage && homePage !== "https://www.iyf.tv" ? `(${homePage})` : ''}
                </button>
            </div>
            <div className="config-info" style={{marginTop: '20px', padding: '15px', backgroundColor: '#f5f5f5', borderRadius: '5px'}}>
                <h3>当前配置:</h3>
                <p><strong>主页:</strong> {homePage || '未设置'}</p>
                <p><strong>代理服务器:</strong> {config.proxyServer || '未设置'}</p>
                <p><strong>用户数据目录:</strong> {config.userData || '默认位置'}</p>
                {config.configPath && (
                    <p><strong>配置文件路径:</strong> <code style={{fontSize: '12px', backgroundColor: '#e0e0e0', padding: '2px 4px', borderRadius: '3px'}}>{config.configPath}</code></p>
                )}
                {homePage && homePage !== "https://www.iyf.tv" && (
                    <p style={{color: '#666', fontSize: '14px'}}>
                        <em>检测到自定义主页，将在1秒后自动跳转...</em>
                    </p>
                )}
            </div>
        </div>
    )
}

export default App

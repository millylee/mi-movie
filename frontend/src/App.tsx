import {useState, useEffect} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Greet, GetHomePage, GetWebViewConfig, GetConfigInfo, SetProxyServer, SetHomePage, SetUserData} from "../wailsjs/go/main/App";

function App() {
    const [homePage, setHomePage] = useState('');
    const [config, setConfig] = useState<any>({});
    const [isLoading, setIsLoading] = useState(true);

    function navigateToURL(url: string) {
        // 直接在当前WebView中导航
        window.location.href = url;
    }

    useEffect(() => {
        // Load configuration and navigate immediately
        Promise.all([
            GetConfigInfo(),
            GetHomePage()
        ]).then(([configData, pageUrl]) => {
            setConfig(configData);
            setHomePage(pageUrl);
            
            // 使用配置的URL或默认URL
            const targetUrl = pageUrl || 'https://www.iyf.tv';
            setIsLoading(false);
            
            // 立即导航到目标URL
            setTimeout(() => {
                navigateToURL(targetUrl);
            }, 100); // 短暂延迟确保页面渲染完成
        }).catch(error => {
            console.error('Failed to load configuration:', error);
            setIsLoading(false);
            // 即使配置加载失败，也导航到默认页面
            setTimeout(() => {
                navigateToURL('https://www.iyf.tv');
            }, 100);
        });
    }, []);

    // 如果正在加载配置，显示加载界面
    if (isLoading) {
        return (
            <div style={{
                width: '100vw', 
                height: '100vh', 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'center',
                backgroundColor: '#f5f5f5'
            }}>
                <div style={{textAlign: 'center'}}>
                    <img src={logo} style={{width: '64px', height: '64px', marginBottom: '20px'}} alt="logo"/>
                    <p style={{fontSize: '18px', color: '#666'}}>正在加载配置...</p>
                </div>
            </div>
        );
    }

    // 主要界面：简洁的加载提示
    return (
        <div style={{width: '100vw', height: '100vh', position: 'relative'}}>
            {/* 如果没有成功导航，显示提示信息 */}
            <div style={{
                width: '100%', 
                height: '100%', 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'center',
                backgroundColor: '#f9f9f9'
            }}>
                <div style={{textAlign: 'center', color: '#666'}}>
                    <p style={{fontSize: '18px', marginBottom: '10px'}}>正在加载网页...</p>
                    <p style={{fontSize: '14px'}}>目标地址: {homePage || 'https://www.iyf.tv'}</p>
                    <p style={{fontSize: '12px', marginTop: '20px', color: '#999'}}>
                        提示: 使用菜单栏 文件 → 设置 来配置应用
                    </p>
                </div>
            </div>
        </div>
    )
}

export default App

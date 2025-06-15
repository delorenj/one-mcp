/**
 * 剪贴板工具函数
 * 提供现代 navigator.clipboard API 和传统 execCommand 的降级方案
 */

export interface CopyResult {
    success: boolean;
    error?: string;
    method?: 'modern' | 'legacy' | 'manual';
}

/**
 * 复制文本到剪贴板
 * @param text 要复制的文本
 * @returns Promise<CopyResult> 复制结果
 */
export async function copyToClipboard(text: string): Promise<CopyResult> {
    // 检查是否支持现代剪贴板 API
    if (navigator.clipboard && window.isSecureContext) {
        try {
            await navigator.clipboard.writeText(text);
            return { success: true, method: 'modern' };
        } catch (error) {
            console.warn('Modern clipboard API failed:', error);
            // 降级到传统方法
            return fallbackCopyToClipboard(text);
        }
    }

    // 降级到传统方法
    return fallbackCopyToClipboard(text);
}

/**
 * 传统的复制方法（降级方案）
 * @param text 要复制的文本
 * @returns CopyResult 复制结果
 */
function fallbackCopyToClipboard(text: string): CopyResult {
    try {
        // 创建临时文本区域
        const textArea = document.createElement('textarea');
        textArea.value = text;

        // 设置样式使其不可见
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        textArea.style.opacity = '0';
        textArea.style.pointerEvents = 'none';

        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        // 尝试执行复制命令
        const successful = document.execCommand('copy');
        document.body.removeChild(textArea);

        if (successful) {
            return { success: true, method: 'legacy' };
        } else {
            return {
                success: false,
                error: 'execCommand_failed',
                method: 'manual'
            };
        }
    } catch (error) {
        return {
            success: false,
            error: 'clipboard_not_supported',
            method: 'manual'
        };
    }
}

/**
 * 检查剪贴板是否可用
 * @returns boolean 是否支持剪贴板操作
 */
export function isClipboardSupported(): boolean {
    return !!(navigator.clipboard && window.isSecureContext) ||
        document.queryCommandSupported?.('copy') === true;
}

/**
 * 获取剪贴板错误的用户友好提示
 * @param error 错误类型
 * @returns string 错误提示键
 */
export function getClipboardErrorMessage(error?: string): string {
    switch (error) {
        case 'execCommand_failed':
            return 'clipboardError.execCommandFailed';
        case 'clipboard_not_supported':
            return 'clipboardError.notSupported';
        default:
            return 'clipboardError.accessDenied';
    }
} 
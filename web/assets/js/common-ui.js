// Common UI components for MeloGo
// 提示框、确认框等公共组件

// 2. Message notification
function showMessage(message, type = 'info', duration = 3000) {
    // Create message element
    const messageEl = document.createElement('div');
    messageEl.className = `alert alert-${type} alert-toast`;
    messageEl.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        z-index: 10000;
        min-width: 200px;
        max-width: 400px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        animation: slideInRight 0.3s ease-out;
    `;
    
    // Add message content
    messageEl.innerHTML = `
        <div style="display: flex; align-items: center;">
            <span style="flex-grow: 1;">${message}</span>
            <button type="button" class="close-toast" style="background: none; border: none; font-size: 1.5em; cursor: pointer; color: inherit; padding: 0 5px;">&times;</button>
        </div>
    `;
    
    // Add to page
    document.body.appendChild(messageEl);
    
    // Add close functionality
    const closeBtn = messageEl.querySelector('.close-toast');
    closeBtn.addEventListener('click', function() {
        if (messageEl.parentNode) {
            messageEl.parentNode.removeChild(messageEl);
        }
    });
    
    // Remove automatically after duration
    if (duration > 0) {
        setTimeout(() => {
            if (messageEl.parentNode) {
                messageEl.parentNode.removeChild(messageEl);
            }
        }, duration);
    }
}

// 3. Alert dialog replacement
function showAlert(message, type = 'info', title = null) {
    // Create modal backdrop
    const backdrop = document.createElement('div');
    backdrop.className = 'alert-modal-backdrop';
    backdrop.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0,0,0,0.5);
        z-index: 10000;
        display: flex;
        justify-content: center;
        align-items: center;
    `;
    
    // Create modal content
    const modal = document.createElement('div');
    modal.className = 'alert-modal';
    modal.style.cssText = `
        background: white;
        border-radius: 8px;
        box-shadow: 0 10px 30px rgba(0,0,0,0.3);
        max-width: 400px;
        width: 90%;
        padding: 20px;
        position: relative;
    `;
    
    // Modal header
    if (title) {
        const header = document.createElement('div');
        header.style.cssText = `
            font-weight: bold;
            font-size: 1.2em;
            margin-bottom: 15px;
            color: var(--text-primary, #333);
        `;
        header.textContent = title;
        modal.appendChild(header);
    }
    
    // Modal body
    const body = document.createElement('div');
    body.style.cssText = `
        margin-bottom: 20px;
        color: var(--text-secondary, #666);
    `;
    body.textContent = message;
    modal.appendChild(body);
    
    // Modal footer
    const footer = document.createElement('div');
    footer.style.cssText = `
        text-align: right;
    `;
    
    const okBtn = document.createElement('button');
    okBtn.className = 'btn btn-primary';
    okBtn.style.cssText = `
        padding: 8px 20px;
        border: none;
        border-radius: 4px;
        background: var(--primary-color, #007bff);
        color: white;
        cursor: pointer;
    `;
    okBtn.textContent = t('ok') || 'OK';
    
    okBtn.addEventListener('click', function() {
        document.body.removeChild(backdrop);
    });
    
    footer.appendChild(okBtn);
    modal.appendChild(footer);
    backdrop.appendChild(modal);
    document.body.appendChild(backdrop);
    
    // Focus the OK button
    setTimeout(() => okBtn.focus(), 100);
    
    // Allow closing with Enter key
    okBtn.addEventListener('keydown', function(e) {
        if (e.key === 'Enter') {
            document.body.removeChild(backdrop);
        }
    });
    
    return new Promise(resolve => {
        okBtn.addEventListener('click', () => resolve(true));
    });
}

// 4. Confirm dialog replacement
function showConfirm(message, title = null) {
    return new Promise((resolve, reject) => {
        // Create modal backdrop
        const backdrop = document.createElement('div');
        backdrop.className = 'confirm-modal-backdrop';
        backdrop.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            z-index: 10000;
            display: flex;
            justify-content: center;
            align-items: center;
        `;
        
        // Create modal content
        const modal = document.createElement('div');
        modal.className = 'confirm-modal';
        modal.style.cssText = `
            background: var(--glass-bg, rgba(15, 23, 42, 0.8));
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border: 1px solid var(--glass-border, rgba(255, 255, 255, 0.1));
            border-radius: var(--radius-xl, 16px);
            box-shadow: var(--shadow-xl, 0 20px 25px -5px rgba(0, 0, 0, 0.2));
            max-width: 400px;
            width: 90%;
            padding: 24px;
            position: relative;
            color: var(--text-primary, #f8fafc);
        `;
        
        // Modal header
        if (title) {
            const header = document.createElement('div');
            header.style.cssText = `
                font-weight: 600;
                font-size: 1.2em;
                margin-bottom: 16px;
                color: var(--text-primary, #f8fafc);
            `;
            header.textContent = title;
            modal.appendChild(header);
        }
        
        // Modal body
        const body = document.createElement('div');
        body.style.cssText = `
            margin-bottom: 24px;
            color: var(--text-primary, #cbd5e1);
        `;
        body.textContent = message;
        modal.appendChild(body);
        
        // Modal footer
        const footer = document.createElement('div');
        footer.style.cssText = `
            text-align: right;
            display: flex;
            justify-content: flex-end;
            gap: 12px;
        `;
        
        const cancelBtn = document.createElement('button');
        cancelBtn.className = 'btn btn-secondary';
        cancelBtn.style.cssText = `
            padding: 8px 16px;
            border-radius: var(--radius-full, 50px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            background: rgba(255, 255, 255, 0.1);
            color: var(--text-primary, #e2e8f0);
            cursor: pointer;
            font-weight: 500;
            transition: all 0.2s;
        `;
        cancelBtn.textContent = t('cancel') || 'Cancel';
        
        const confirmBtn = document.createElement('button');
        confirmBtn.className = 'btn btn-danger';
        confirmBtn.style.cssText = `
            padding: 8px 16px;
            border-radius: var(--radius-full, 50px);
            border: none;
            background: rgba(239, 68, 68, 0.2);
            color: #ef4444;
            cursor: pointer;
            font-weight: 500;
            transition: all 0.2s;
        `;
        confirmBtn.textContent = t('confirm') || 'Confirm';
        
        // Add hover effects to match theme
        cancelBtn.addEventListener('mouseenter', () => {
            cancelBtn.style.background = 'rgba(255, 255, 255, 0.2)';
        });
        
        cancelBtn.addEventListener('mouseleave', () => {
            cancelBtn.style.background = 'rgba(255, 255, 255, 0.1)';
        });
        
        confirmBtn.addEventListener('mouseenter', () => {
            confirmBtn.style.background = '#ef4444';
            confirmBtn.style.color = 'white';
        });
        
        confirmBtn.addEventListener('mouseleave', () => {
            confirmBtn.style.background = 'rgba(239, 68, 68, 0.2)';
            confirmBtn.style.color = '#ef4444';
        });
        
        cancelBtn.addEventListener('click', function() {
            document.body.removeChild(backdrop);
            resolve(false);
        });
        
        confirmBtn.addEventListener('click', function() {
            document.body.removeChild(backdrop);
            resolve(true);
        });
        
        footer.appendChild(cancelBtn);
        footer.appendChild(confirmBtn);
        modal.appendChild(footer);
        backdrop.appendChild(modal);
        document.body.appendChild(backdrop);
        
        // Focus the confirm button
        setTimeout(() => confirmBtn.focus(), 100);
        
        // Allow closing with Escape key
        document.addEventListener('keydown', function escHandler(e) {
            if (e.key === 'Escape') {
                document.body.removeChild(backdrop);
                resolve(false);
                document.removeEventListener('keydown', escHandler);
            }
        });
    });
}

// 5. Update showAlert function in login/register pages to use the new modal
function showAlertInContainer(message, type, containerId) {
    const container = document.getElementById(containerId);
    if (container) {
        container.innerHTML = `
            <div class="alert alert-${type} alert-dismissible fade show" role="alert">
                ${message}
                <button type="button" class="close" data-dismiss="alert" aria-label="Close">
                    <span aria-hidden="true">&times;</span>
                </button>
            </div>
        `;
        
        // Auto-hide after 5 seconds
        setTimeout(() => {
            if (container.innerHTML) {
                container.innerHTML = '';
            }
        }, 5000);
    } else {
        // Fallback to the modal if container not found
        showMessage(message, type);
    }
}

// 6. Add CSS for animations (if not already present)
function addCommonStyles() {
    const styleId = 'common-ui-styles';
    if (document.getElementById(styleId)) return; // Already added
    
    const style = document.createElement('style');
    style.id = styleId;
    style.textContent = `
        @keyframes slideInRight {
            from {
                transform: translateX(100%);
                opacity: 0;
            }
            to {
                transform: translateX(0);
                opacity: 1;
            }
        }
        
        .alert-toast {
            animation: slideInRight 0.3s ease-out;
        }
    `;
    document.head.appendChild(style);
}

// Initialize styles when DOM is loaded
document.addEventListener('DOMContentLoaded', addCommonStyles);
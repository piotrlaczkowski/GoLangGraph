// GoLangGraph Documentation Interactive Features

document.addEventListener('DOMContentLoaded', function() {
    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Add loading animation to external links
    document.querySelectorAll('a[href^="http"]').forEach(link => {
        link.addEventListener('click', function() {
            this.innerHTML += ' <span class="loading-spinner">⏳</span>';
            setTimeout(() => {
                const spinner = this.querySelector('.loading-spinner');
                if (spinner) {
                    spinner.remove();
                }
            }, 2000);
        });
    });

    // Animate cards on scroll
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    // Observe all cards
    document.querySelectorAll('.md-typeset .grid.cards > *').forEach(card => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(20px)';
        card.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(card);
    });

    // Add copy button to code blocks
    document.querySelectorAll('.highlight').forEach(block => {
        const button = document.createElement('button');
        button.className = 'copy-button';
        button.innerHTML = '📋 Copy';
        button.style.cssText = `
            position: absolute;
            top: 8px;
            right: 8px;
            background: var(--golanggraph-gradient);
            color: white;
            border: none;
            border-radius: 4px;
            padding: 4px 8px;
            font-size: 12px;
            cursor: pointer;
            opacity: 0;
            transition: opacity 0.3s ease;
        `;
        
        block.style.position = 'relative';
        block.appendChild(button);
        
        block.addEventListener('mouseenter', () => {
            button.style.opacity = '1';
        });
        
        block.addEventListener('mouseleave', () => {
            button.style.opacity = '0';
        });
        
        button.addEventListener('click', () => {
            const code = block.querySelector('code');
            if (code) {
                navigator.clipboard.writeText(code.textContent).then(() => {
                    button.innerHTML = '✅ Copied!';
                    setTimeout(() => {
                        button.innerHTML = '📋 Copy';
                    }, 2000);
                });
            }
        });
    });

    // Add progress bar for page reading
    const progressBar = document.createElement('div');
    progressBar.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 0%;
        height: 3px;
        background: var(--golanggraph-gradient);
        z-index: 9999;
        transition: width 0.3s ease;
    `;
    document.body.appendChild(progressBar);

    window.addEventListener('scroll', () => {
        const scrollTop = window.pageYOffset;
        const docHeight = document.body.scrollHeight - window.innerHeight;
        const scrollPercent = (scrollTop / docHeight) * 100;
        progressBar.style.width = scrollPercent + '%';
    });

    // Add keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        // Press 'S' to focus search
        if (e.key === 's' && !e.ctrlKey && !e.metaKey) {
            const searchInput = document.querySelector('.md-search__input');
            if (searchInput && document.activeElement !== searchInput) {
                e.preventDefault();
                searchInput.focus();
            }
        }
        
        // Press 'T' to toggle theme
        if (e.key === 't' && !e.ctrlKey && !e.metaKey) {
            const themeToggle = document.querySelector('[data-md-component="palette"]');
            if (themeToggle) {
                e.preventDefault();
                themeToggle.click();
            }
        }
    });

    // Add tooltips to navigation items
    document.querySelectorAll('.md-nav__link').forEach(link => {
        const title = link.getAttribute('title') || link.textContent;
        link.setAttribute('title', title);
    });

    // Enhance performance metrics display
    document.querySelectorAll('.performance-metric').forEach(metric => {
        metric.addEventListener('mouseenter', function() {
            this.style.transform = 'scale(1.05)';
            this.style.transition = 'transform 0.3s ease';
        });
        
        metric.addEventListener('mouseleave', function() {
            this.style.transform = 'scale(1)';
        });
    });

    // Add GitHub star button animation
    document.querySelectorAll('a[href*="github.com"]').forEach(link => {
        if (link.textContent.includes('Star')) {
            link.addEventListener('click', function() {
                this.innerHTML += ' ⭐';
                setTimeout(() => {
                    this.innerHTML = this.innerHTML.replace(' ⭐', '');
                }, 3000);
            });
        }
    });

    // Add version indicator
    const versionIndicator = document.createElement('div');
    versionIndicator.innerHTML = '🚀 v1.0.0';
    versionIndicator.style.cssText = `
        position: fixed;
        bottom: 20px;
        right: 20px;
        background: var(--golanggraph-gradient);
        color: white;
        padding: 8px 12px;
        border-radius: 20px;
        font-size: 12px;
        font-weight: 600;
        z-index: 1000;
        opacity: 0.8;
        transition: opacity 0.3s ease;
    `;
    
    versionIndicator.addEventListener('mouseenter', () => {
        versionIndicator.style.opacity = '1';
    });
    
    versionIndicator.addEventListener('mouseleave', () => {
        versionIndicator.style.opacity = '0.8';
    });
    
    document.body.appendChild(versionIndicator);

    // Add easter egg - Konami code
    let konamiCode = [];
    const konamiSequence = [
        'ArrowUp', 'ArrowUp', 'ArrowDown', 'ArrowDown',
        'ArrowLeft', 'ArrowRight', 'ArrowLeft', 'ArrowRight',
        'KeyB', 'KeyA'
    ];

    document.addEventListener('keydown', (e) => {
        konamiCode.push(e.code);
        if (konamiCode.length > konamiSequence.length) {
            konamiCode.shift();
        }
        
        if (konamiCode.join(',') === konamiSequence.join(',')) {
            // Easter egg activated!
            document.body.style.animation = 'rainbow 2s infinite';
            setTimeout(() => {
                document.body.style.animation = '';
            }, 5000);
            
            // Add rainbow animation
            const style = document.createElement('style');
            style.textContent = `
                @keyframes rainbow {
                    0% { filter: hue-rotate(0deg); }
                    100% { filter: hue-rotate(360deg); }
                }
            `;
            document.head.appendChild(style);
            
            konamiCode = [];
        }
    });

    // Add scroll-to-top button
    const scrollToTopBtn = document.createElement('button');
    scrollToTopBtn.innerHTML = '↑';
    scrollToTopBtn.style.cssText = `
        position: fixed;
        bottom: 80px;
        right: 20px;
        width: 40px;
        height: 40px;
        border-radius: 50%;
        background: var(--golanggraph-gradient);
        color: white;
        border: none;
        font-size: 18px;
        cursor: pointer;
        opacity: 0;
        transition: opacity 0.3s ease, transform 0.3s ease;
        z-index: 1000;
    `;
    
    scrollToTopBtn.addEventListener('click', () => {
        window.scrollTo({
            top: 0,
            behavior: 'smooth'
        });
    });
    
    scrollToTopBtn.addEventListener('mouseenter', () => {
        scrollToTopBtn.style.transform = 'scale(1.1)';
    });
    
    scrollToTopBtn.addEventListener('mouseleave', () => {
        scrollToTopBtn.style.transform = 'scale(1)';
    });
    
    document.body.appendChild(scrollToTopBtn);
    
    window.addEventListener('scroll', () => {
        if (window.pageYOffset > 300) {
            scrollToTopBtn.style.opacity = '1';
        } else {
            scrollToTopBtn.style.opacity = '0';
        }
    });

    console.log('🚀 GoLangGraph Documentation loaded successfully!');
    console.log('💡 Tip: Press "S" to search, "T" to toggle theme');
}); 
 document.addEventListener('DOMContentLoaded', function() {
            const flashMessage = document.getElementById('flash-message');
            if (flashMessage) {
             
                setTimeout(() => {
                    flashMessage.classList.add('opacity-0', 'transition-opacity', 'duration-500');
                    flashMessage.addEventListener('transitionend', () => flashMessage.remove());
                }, 5000); 

           
                const urlParams = new URLSearchParams(window.location.search);
                if (urlParams.has('status') || urlParams.has('message')) {
                    urlParams.delete('status');
                    urlParams.delete('message');
                    const newUrl = window.location.pathname + (urlParams.toString() ? '?' + urlParams.toString() : '');
                    window.history.replaceState({}, document.title, newUrl);
                }
            }

        
            const closeButton = flashMessage ? flashMessage.querySelector('button') : null;
            if (closeButton) {
                closeButton.addEventListener('click', function() {
                    flashMessage.remove(); 
                   
                    const urlParams = new URLSearchParams(window.location.search);
                    if (urlParams.has('status') || urlParams.has('message')) {
                        urlParams.delete('status');
                        urlParams.delete('message');
                        const newUrl = window.location.pathname + (urlParams.toString() ? '?' + urlParams.toString() : '');
                        window.history.replaceState({}, document.title, newUrl);
                    }
                });
            }
        });
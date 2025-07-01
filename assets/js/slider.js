let currentSlide = 0;
const slides = document.querySelectorAll('#slider .slides img');

function showSlide(index) {
  slides.forEach((slide, i) => {
    slide.classList.remove('opacity-100');
    slide.classList.add('opacity-0');
    if (i === index) {
      slide.classList.add('opacity-100');
      slide.classList.remove('opacity-0');
    }
  });
  currentSlide = index;
}

function nextSlide() {
  const nextIndex = (currentSlide + 1) % slides.length;
  showSlide(nextIndex);
}


showSlide(currentSlide);
setInterval(nextSlide, 5000);

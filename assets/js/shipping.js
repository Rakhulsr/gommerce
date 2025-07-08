
function formatCurrency(amount) {
  
    const num = parseInt(amount);
    if (isNaN(num)) {
        return "Rp 0"; 
    }
    return `Rp ${num.toLocaleString('id-ID')}`;
}

document.addEventListener('DOMContentLoaded', function() {
    const provinceSelect = document.getElementById('province_id');
    const citySelect = document.getElementById('city_id');
    const courierSelect = document.getElementById('courier');
    const shippingFeeSelect = document.getElementById('shipping_fee');
    const calculateShippingForm = document.getElementById('calculate-shipping');
    const shippingMessage = document.getElementById('shiping-calculation-msg');
    
    const shippingFeeDisplay = document.getElementById('shipping-fee-display'); 
    const grandTotalDisplay = document.getElementById('grand-total-display');   
    const cartSubtotalInput = document.getElementById('cart_subtotal_input');   

    const totalWeightInput = document.getElementById('total_weight_input');
    let totalWeight = 1; 
    if (totalWeightInput && !isNaN(parseInt(totalWeightInput.value))) {
        totalWeight = parseInt(totalWeightInput.value);
        if (totalWeight <= 0) {
            totalWeight = 1;
        }
    }

    const originCityIDInput = document.getElementById('origin_city_id_input');
    const originCityID = originCityIDInput ? String(originCityIDInput.value) : ""; 

    const APP_API_BASE_URL = "/"; 

    let cartSubtotal = 0; 
    if (cartSubtotalInput && !isNaN(parseInt(cartSubtotalInput.value))) {
        cartSubtotal = parseInt(cartSubtotalInput.value);
    }

    function updateGrandTotalDisplay(selectedShippingCostValue = "0") {
        const currentShippingFee = parseInt(selectedShippingCostValue);
        const grandTotalCalculated = cartSubtotal + currentShippingFee;

        shippingFeeDisplay.textContent = formatCurrency(currentShippingFee);
        grandTotalDisplay.textContent = formatCurrency(grandTotalCalculated);
    }

  
    updateGrandTotalDisplay(shippingFeeSelect.value); 


    async function loadCities(provinceId) {
        citySelect.innerHTML = '<option value="" selected>--Memuat Kota/Kabupaten--</option>';
        citySelect.disabled = true;
        shippingFeeSelect.innerHTML = '<option value="" selected>--Pilih Opsi Pengiriman--</option>';
        shippingFeeSelect.disabled = true;
        shippingMessage.textContent = ''; 
        shippingMessage.classList.remove('text-red-500', 'text-yellow-500', 'text-green-500');
        updateGrandTotalDisplay(0); 

        if (!provinceId) {
            citySelect.innerHTML = '<option value="" selected>--Pilih Kota/Kabupaten--</option>';
            citySelect.disabled = false;
            return;
        }

        try {
            const url = `${APP_API_BASE_URL}cities?province_id=${provinceId}`;
            console.log("Fetching cities from URL:", url); 
            
            const response = await fetch(url, { 
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            const contentType = response.headers.get("content-type");
            if (!contentType || !contentType.includes("application/json")) {
                const errorText = await response.text();
                console.error('Expected JSON response for cities, but received:', errorText);
                shippingMessage.textContent = 'Gagal memuat kota: Respons tidak valid atau bukan JSON.';
                shippingMessage.classList.add('text-red-500');
                citySelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
                return;
            }

            if (!response.ok) { 
                const errorData = await response.json().catch(() => ({})); 
                console.error('HTTP Error fetching cities:', response.status, errorData);
                shippingMessage.textContent = `Gagal memuat kota: ${errorData.message || response.statusText}`;
                shippingMessage.classList.add('text-red-500');
                citySelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
                return;
            }

            const data = await response.json(); 

            if (data.cities) { 
                citySelect.innerHTML = '<option value="" selected>--Pilih Kota/Kabupaten--</option>';
                data.cities.forEach(city => {
                    const option = document.createElement('option');
                    option.value = city.city_id; 
                    option.textContent = `${city.type} ${city.city_name}`; 
                    citySelect.appendChild(option);
                });
                citySelect.disabled = false;
            } else {
                console.error('Backend response for cities missing "cities" property:', data);
                shippingMessage.textContent = 'Gagal memuat kota: Data tidak lengkap.';
                shippingMessage.classList.add('text-red-500');
                citySelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
            }
        } catch (error) {
            console.error('Network error fetching cities from backend:', error);
            shippingMessage.textContent = 'Terjadi kesalahan jaringan saat memuat kota.';
            shippingMessage.classList.add('text-red-500');
            citySelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
        } finally {
            citySelect.disabled = false;
        }
    }

    async function loadShippingCosts() {
        shippingFeeSelect.innerHTML = '<option value="" selected>--Memuat Opsi Pengiriman--</option>';
        shippingFeeSelect.disabled = true;
        shippingMessage.textContent = ''; 
        shippingMessage.classList.remove('text-red-500', 'text-yellow-500', 'text-green-500');
        updateGrandTotalDisplay(0); 

        const destinationCityID = citySelect.value;
        const courier = courierSelect.value;
        
        if (!originCityID) {
            shippingMessage.textContent = 'Kota asal pengiriman belum ditentukan.';
            shippingMessage.classList.add('text-red-500');
            shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat (Asal)--</option>';
            return; 
        }

        if (!destinationCityID || !courier || totalWeight <= 0) {
            shippingMessage.textContent = 'Mohon lengkapi pilihan Kota/Kabupaten dan Kurir, serta pastikan berat lebih dari 0.';
            shippingMessage.classList.add('text-yellow-500');
            shippingFeeSelect.innerHTML = '<option value="" selected>--Pilih Opsi Pengiriman--</option>';
            shippingFeeSelect.disabled = false;
            return;
        }

        try {
            const url = `${APP_API_BASE_URL}calculate-shipping-cost`;
            console.log("Fetching shipping costs from URL:", url); 
            
            const response = await fetch(url, { 
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json' 
                },
                body: JSON.stringify({ 
                    origin: String(originCityID),
                    destination: String(destinationCityID),
                    weight: totalWeight,
                    courier: courier
                })
            });

            const contentType = response.headers.get("content-type");
            if (!contentType || !contentType.includes("application/json")) {
                const errorText = await response.text();
                console.error('Expected JSON response for shipping costs, but received:', errorText);
                shippingMessage.textContent = 'Gagal menghitung ongkir: Respons tidak valid atau bukan JSON.';
                shippingMessage.classList.add('text-red-500');
                shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
                return;
            }

            if (!response.ok) { 
                const errorData = await response.json().catch(() => ({})); 
                console.error('HTTP Error fetching shipping costs:', response.status, errorData);
                let errorMessage = `Gagal menghitung ongkir: ${errorData.message || response.statusText}`;
                if (errorData && errorData.error) { 
                    errorMessage = `Gagal menghitung ongkir: ${errorData.error}`;
                } else if (errorData && errorData.rajaongkir && errorData.rajaongkir.status && errorData.rajaongkir.status.description) {
                    errorMessage = `Gagal menghitung ongkir: ${errorData.rajaongkir.status.description}`;
                }
                shippingMessage.textContent = errorMessage;
                shippingMessage.classList.add('text-red-500');
                shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
                return;
            }

            const data = await response.json(); 

            if (data.costs) { 
                shippingFeeSelect.innerHTML = '<option value="" selected>--Pilih Opsi Pengiriman--</option>';
                const results = data.costs;
                if (results.length > 0 && results[0].costs.length > 0) {
                    results[0].costs.forEach(cost => {
                        cost.cost.forEach(service => {
                            const option = document.createElement('option');
                            const formattedValue = `${results[0].code.toUpperCase()} - ${cost.service} (${service.etd} hari): ${formatCurrency(service.value)}`;
                            option.value = service.value; 
                            option.textContent = formattedValue;
                            shippingFeeSelect.appendChild(option);
                        });
                    });
                    shippingFeeSelect.disabled = false;
                    shippingMessage.textContent = ''; 
                    shippingMessage.classList.remove('text-red-500', 'text-yellow-500', 'text-green-500');
                } else {
                    shippingMessage.textContent = 'Tidak ada opsi pengiriman tersedia untuk rute ini.';
                    shippingMessage.classList.add('text-yellow-500');
                    shippingFeeSelect.innerHTML = '<option value="" selected>--Tidak Tersedia--</option>';
                }
            } else {
                console.error('Backend response for shipping costs missing "costs" property:', data);
                let errorMessage = 'Gagal menghitung ongkir: Data tidak lengkap.';
                if (data && data.error) {
                    errorMessage = `Gagal menghitung ongkir: ${data.error}`;
                } else if (data && data.rajaongkir && data.rajaongkir.status && data.rajaongkir.status.description) {
                    errorMessage = `Gagal menghitung ongkir: ${data.rajaongkir.status.description}`;
                }
                shippingMessage.textContent = errorMessage;
                shippingMessage.classList.add('text-red-500');
                shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
            }
        } catch (error) {
            console.error('Network error fetching shipping costs from backend:', error);
            shippingMessage.textContent = `Terjadi kesalahan jaringan: ${error.message}.`;
            shippingMessage.classList.add('text-red-500');
            shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
        } finally {
            shippingFeeSelect.disabled = false;
        }
    }
    
    provinceSelect.addEventListener('change', function() {
        loadCities(this.value);
    });

    citySelect.addEventListener('change', function() {
        loadShippingCosts();
    });

    courierSelect.addEventListener('change', function() {
        loadShippingCosts();
    });

    shippingFeeSelect.addEventListener('change', function() {
        const selectedValue = this.value; 
        updateGrandTotalDisplay(selectedValue);
        
        if (selectedValue !== "") {
            shippingMessage.textContent = `Opsi pengiriman dipilih: ${formatCurrency(selectedValue)}`;
            shippingMessage.classList.remove('text-red-500', 'text-yellow-500');
            shippingMessage.classList.add('text-green-500');
        } else {
            shippingMessage.textContent = 'Silakan pilih opsi pengiriman.';
            shippingMessage.classList.remove('text-green-500');
            shippingMessage.classList.add('text-red-500');
        }
    });

    calculateShippingForm.addEventListener('submit', function(event) {
        event.preventDefault(); 
        const selectedShippingCost = shippingFeeSelect.value;
        
        if (selectedShippingCost && selectedShippingCost !== "") { 
            const shippingCost = parseInt(selectedShippingCost);
            const finalPrice = cartSubtotal + shippingCost; 
            
            console.log('Biaya pengiriman yang dipilih:', shippingCost);
            console.log('Grand Total Akhir (Subtotal + Ongkir):', finalPrice);


            shippingFeeDisplay.textContent = formatCurrency(shippingCost);
            grandTotalDisplay.textContent = formatCurrency(finalPrice);

            shippingMessage.textContent = `Pesanan siap diproses dengan biaya pengiriman: ${formatCurrency(shippingCost)}. Grand Total: ${formatCurrency(finalPrice)}`;
            shippingMessage.classList.remove('text-red-500', 'text-yellow-500');
            shippingMessage.classList.add('text-green-500');
            
        } else {
            shippingMessage.textContent = 'Silakan pilih opsi pengiriman sebelum melanjutkan.';
            shippingMessage.classList.add('text-red-500');
        }
    });

    
    if (provinceSelect.value) {
        loadCities(provinceSelect.value);
    }
});
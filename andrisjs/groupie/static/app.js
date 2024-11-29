// Function to fetch band data from the server
async function fetchBands() {
  const response = await fetch('/api/bands');
  if (!response.ok) {
    console.error('Failed to fetch band data:', response.statusText);
    return [];
  }
  const data = await response.json();
  return data;
}

// Function to dynamically create cards for each band
function createCards(bands) {
  const bandContainer = document.getElementById('band-container');
  bandContainer.innerHTML = ''; // Clear any existing content

  bands.forEach((band, index) => {
    const card = document.createElement('div');
    card.className = 'card';
    card.innerHTML = `
      <img src="${band.image}" alt="${band.name}">
      <div class="name">${band.name}</div>
      <div class="album">${band.album} (${band.year})</div>
    `;
    card.addEventListener('click', () => expandCard(band));
    bandContainer.appendChild(card);
  });
}

// Function to expand a card and show band details
function expandCard(band) {
  // Create an overlay to dim the background
  const overlay = document.createElement('div');
  overlay.className = 'overlay';
  overlay.addEventListener('click', closeExpandedView);

  // Create the expanded card
  const expandedCard = document.createElement('div');
  expandedCard.className = 'expanded-card';
  expandedCard.innerHTML = `
    <div class="sidebar-left" style="background-image: url('${band.image}');"></div>
    <div class="content">
      <h2>${band.name}</h2>
      <p><strong>Year:</strong> ${band.year}</p>
      <p><strong>First Album:</strong> ${band.album}</p>
      <p><strong>Members:</strong> ${band.members.join(', ')}</p>
      <h3>Upcoming Dates and Locations</h3>
      <ul>
        ${band.dates.map((date, i) => `<li>${date} - ${band.locations[i]}</li>`).join('')}
      </ul>
    </div>
    <div class="sidebar-right" style="background-image: url('${band.image}');"></div>
    <button class="close-btn">Close</button>
  `;

  // Add close button functionality
  expandedCard.querySelector('.close-btn').addEventListener('click', closeExpandedView);

  // Append overlay and expanded card to the document
  document.body.appendChild(overlay);
  document.body.appendChild(expandedCard);

  // Add animation to expand the card
  setTimeout(() => expandedCard.classList.add('expanded'), 10); // Add animation after a slight delay
}

// Function to close the expanded view
function closeExpandedView() {
  const overlay = document.querySelector('.overlay');
  const expandedCard = document.querySelector('.expanded-card');
  if (overlay) overlay.remove();
  if (expandedCard) expandedCard.remove();
}

// Initialize the app
async function init() {
  const bands = await fetchBands(); // Fetch band data from the API
  createCards(bands);              // Create and display cards
}

// Run the app on load
init();
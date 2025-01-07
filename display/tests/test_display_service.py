import os
import platform
from unittest.mock import MagicMock, patch
import pytest
from PIL import Image

from ..display_service import DisplayService

@pytest.fixture
def temp_image(tmp_path):
    """Create a temporary test image."""
    image_path = tmp_path / "test_image.png"
    img = Image.new('RGB', (100, 100), color='red')
    img.save(image_path)
    return str(image_path)

@pytest.fixture
def mock_inky():
    """Mock the Inky display."""
    with patch('inky.auto.auto') as mock:
        display = MagicMock()
        mock.return_value = display
        yield display

def test_init_non_pi():
    """Test initialization on non-Pi platform."""
    with patch('platform.system', return_value='Darwin'), \
         patch('platform.machine', return_value='x86_64'):
        service = DisplayService()
        assert service.width == 800
        assert service.height == 480
        assert not service.is_pi
        assert service.display is None

def test_init_pi(mock_inky):
    """Test initialization on Pi platform."""
    with patch('platform.system', return_value='Linux'), \
         patch('platform.machine', return_value='aarch64'):
        service = DisplayService()
        assert service.width == 800
        assert service.height == 480
        assert service.is_pi
        assert service.display == mock_inky

@pytest.mark.asyncio
async def test_update_display_non_pi(temp_image):
    """Test display update on non-Pi platform."""
    with patch('platform.system', return_value='Darwin'), \
         patch('platform.machine', return_value='x86_64'):
        service = DisplayService()
        await service.update_display(temp_image)
        # Verify image was resized
        img = Image.open(temp_image)
        assert img.size == (800, 480)

@pytest.mark.asyncio
async def test_update_display_pi(temp_image, mock_inky):
    """Test display update on Pi platform."""
    with patch('platform.system', return_value='Linux'), \
         patch('platform.machine', return_value='aarch64'):
        service = DisplayService()
        await service.update_display(temp_image)
        # Verify display methods were called
        mock_inky.set_image.assert_called_once()
        mock_inky.show.assert_called_once()
        # Verify image was resized
        img = Image.open(temp_image)
        assert img.size == (800, 480)
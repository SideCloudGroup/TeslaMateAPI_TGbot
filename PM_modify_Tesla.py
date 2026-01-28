from pagermaid import log, bot
from pagermaid.listener import listener
from pagermaid.utils import alias_command
from requests import get

version = "0.1"
API_BASE = "https://example.com"  # 请替换为你的TeslaMate API URL
HEADERS = {
    "CF-Access-Client-Id": "",  # 请替换为你的Client-Id
    "CF-Access-Client-Secret": ""  # 请替换为你的Client-Secret
}


async def get_cars():
    try:
        response = get(f"{API_BASE}/api/v1/cars", headers=HEADERS)
        if response.status_code != 200:
            return f"获取车辆信息失败，状态码: {response.status_code}"
        data = response.json()
        cars = data['data']['cars']
        result = "🚗 车辆列表：\n"
        for car in cars:
            result += f"ID: {car['car_id']}\n名称: {car['name']}\n型号: {car['car_details']['model']}\n颜色: {car['car_exterior']['exterior_color']}\n\n"
        return result.strip()
    except Exception as e:
        await log(f"获取车辆信息时发生错误: {e}")
        return "获取车辆信息时发生错误"

async def get_car_id():
    try:
        response = get(f"{API_BASE}/api/v1/cars", headers=HEADERS)
        if response.status_code != 200:
            return None
        data = response.json()
        return data['data']['cars'][0]['car_id'] if data['data']['cars'] else None
    except Exception as e:
        await log(f"获取车辆ID时发生错误: {e}")
        return None

async def get_status():
    car_id = await get_car_id()
    if not car_id:
        return "无法获取车辆ID"
    try:
        response = get(f"{API_BASE}/api/v1/cars/{car_id}/status", headers=HEADERS)
        if response.status_code != 200:
            return f"获取状态失败，状态码: {response.status_code}"
        data = response.json()
        status = data['data']['status']
        car_status = status['car_status']
        battery = status['battery_details']
        charging = status['charging_details']
        climate = status['climate_details']
        result = f"🚗 车辆: {status['display_name']}\n📍 状态: {status['state']}\n⏰ 自上次状态以来: {status['state_since']}\n📏 里程表: {status['odometer']} km\n\n🔒 车门状态:\n锁定: {'是' if car_status['locked'] else '否'}\n哨兵模式: {'开' if car_status['sentry_mode'] else '关'}\n车窗: {'开' if car_status['windows_open'] else '关'}\n车门: {'开' if car_status['doors_open'] else '关'}\n后备箱: {'开' if car_status['trunk_open'] else '关'}\n前备箱: {'开' if car_status['frunk_open'] else '关'}\n\n🔋 电池:\n电量: {battery['battery_level']}%\n预计续航: {battery['est_battery_range']} km\n额定续航: {battery['rated_battery_range']} km\n\n⚡ 充电:\n充电状态: {charging['charging_state']}\n充电限制: {charging['charge_limit_soc']}%\n\n🌡️ 气候:\n空调: {'开' if climate['is_climate_on'] else '关'}\n内部温度: {climate['inside_temp']}°C\n外部温度: {climate['outside_temp']}°C\n\n📱 版本:\n版本号: {status['car_versions']['version']}\n更新可用: {'是' if status['car_versions']['update_available'] else '否'}\n"
        return result.strip()
    except Exception as e:
        await log(f"获取状态时发生错误: {e}")
        return "获取状态时发生错误"

async def get_charges():
    car_id = await get_car_id()
    if not car_id:
        return "无法获取车辆ID"
    try:
        response = get(f"{API_BASE}/api/v1/cars/{car_id}/charges", headers=HEADERS)
        if response.status_code != 200:
            return f"获取充电记录失败，状态码: {response.status_code}"
        data = response.json()
        charges = data['data']['charges']
        if not charges:
            return "无充电记录"
        result = "🔌 最近充电记录：\n"
        for charge in charges[:5]:  # 显示最近5条
            result += f"开始: {charge['start_date']}\n结束: {charge['end_date']}\n地址: {charge['address']}\n增加电量: {charge['charge_energy_added']} kWh\n费用: {charge['cost']}\n持续时间: {charge['duration_str']}\n电池变化: {charge['battery_details']['start_battery_level']}% -> {charge['battery_details']['end_battery_level']}%\n\n"
        return result.strip()
    except Exception as e:
        await log(f"获取充电记录时发生错误: {e}")
        return "获取充电记录时发生错误"

async def get_drives():
    car_id = await get_car_id()
    if not car_id:
        return "无法获取车辆ID"
    try:
        response = get(f"{API_BASE}/api/v1/cars/{car_id}/drives", headers=HEADERS)
        if response.status_code != 200:
            return f"获取驾驶记录失败，状态码: {response.status_code}"
        data = response.json()
        drives = data['data']['drives']
        if not drives:
            return "无驾驶记录"
        result = "🚗 最近驾驶记录：\n"
        for drive in drives[:5]:  # 显示最近5条
            result += f"开始: {drive['start_date']}\n结束: {drive['end_date']}\n起点: {drive['start_address']}\n终点: {drive['end_address']}\n距离: {drive['odometer_details']['odometer_distance']:.2f} km\n持续时间: {drive['duration_str']}\n平均速度: {drive['speed_avg']:.1f} km/h\n最高速度: {drive['speed_max']} km/h\n电池变化: {drive['battery_details']['start_battery_level']}% -> {drive['battery_details']['end_battery_level']}%\n\n"
        return result.strip()
    except Exception as e:
        await log(f"获取驾驶记录时发生错误: {e}")
        return "获取驾驶记录时发生错误"

@listener(outgoing=True, command=alias_command("tesla"),
          description="TeslaMate API 查询 (使用 -tesla help 查看所有指令)", parameters="<command>")
async def tesla(context):
    command = context.arguments
    if command == "help":
        await context.edit(f"TeslaMate API - V{version}\n"
                           "-tesla cars - 获取车辆列表\n"
                           "-tesla status - 获取车辆状态\n"
                           "-tesla charges - 获取充电记录\n"
                           "-tesla drives - 获取驾驶记录\n"
                           "-tesla help - 获取帮助")
        return
    elif command == "cars":
        await context.edit(await get_cars())
        return
    elif command == "status":
        await context.edit(await get_status())
        return
    elif command == "charges":
        await context.edit(await get_charges())
        return
    elif command == "drives":
        await context.edit(await get_drives())
        return
    else:
        await context.edit("未知指令，请使用 `-tesla help` 查看帮助。")
        return

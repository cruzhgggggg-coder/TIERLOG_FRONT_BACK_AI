<script setup>
import { onMounted, ref, onUnmounted } from 'vue';

const props = defineProps({
    processing: {
        type: Boolean,
        default: false
    }
});

const canvasRef = ref(null);
const mouse = ref({ x: null, y: null });
let animationId = null;

class Node {
    constructor(x, y) {
        this.x = x;
        this.y = y;
        this.baseX = x;
        this.baseY = y;
        this.size = Math.random() * 1.5 + 0.5;
        this.pulse = Math.random() * Math.PI * 2;
        this.density = (Math.random() * 30) + 10;
        this.isFiring = false;
    }

    draw(ctx, isProcessing) {
        const p = Math.sin(this.pulse) * 0.5 + 0.5;
        ctx.beginPath();
        
        // Synaptic firing during processing
        if (isProcessing && Math.random() > 0.995) this.isFiring = true;
        
        const sizeMultiplier = this.isFiring ? 4 : 1;
        ctx.arc(this.x, this.y, (this.size + p * 2) * sizeMultiplier, 0, Math.PI * 2);
        
        const color = isProcessing ? '255, 0, 234' : '0, 242, 255'; 
        const opacity = this.isFiring ? 1 : (0.2 + p * 0.6);
        
        ctx.fillStyle = this.isFiring ? '#ffffff' : `rgba(${color}, ${opacity})`;
        ctx.fill();
        
        ctx.shadowBlur = (this.isFiring ? 30 : 15) * p;
        ctx.shadowColor = this.isFiring ? '#ffffff' : `rgba(${color}, 0.5)`;
        
        if (this.isFiring) {
            setTimeout(() => this.isFiring = false, 100);
        }
        
        this.pulse += isProcessing ? 0.08 : 0.04;
    }

    update() {
        let dx = (mouse.value.x !== null ? mouse.value.x : this.baseX) - this.x;
        let dy = (mouse.value.y !== null ? mouse.value.y : this.baseY) - this.y;
        let distance = Math.sqrt(dx * dx + dy * dy);
        
        if (mouse.value.x !== null && distance < 300) {
            let forceDirectionX = dx / distance;
            let forceDirectionY = dy / distance;
            let force = (300 - distance) / 300;
            this.x += forceDirectionX * force * this.density * 0.4;
            this.y += forceDirectionY * force * this.density * 0.4;
        } else {
            // Return to base with easing
            this.x += (this.baseX - this.x) * 0.05;
            this.y += (this.baseY - this.y) * 0.05;
        }
    }
}

const handleMouseMove = (e) => {
    const canvas = canvasRef.value;
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    mouse.value.x = e.clientX - rect.left;
    mouse.value.y = e.clientY - rect.top;
};

const handleMouseLeave = () => {
    mouse.value.x = null;
    mouse.value.y = null;
};

const initCanvas = () => {
    const canvas = canvasRef.value;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const width = canvas.width = canvas.offsetWidth;
    const height = canvas.height = canvas.offsetHeight;

    const nodes = [];
    const nodeCount = Math.floor((width * height) / 15000); // Responsive density

    for (let i = 0; i < nodeCount; i++) {
        nodes.push(new Node(Math.random() * width, Math.random() * height));
    }

    const animate = () => {
        ctx.clearRect(0, 0, width, height);
        
        ctx.lineWidth = 0.5;
        for (let i = 0; i < nodes.length; i++) {
            for (let j = i + 1; j < nodes.length; j++) {
                const dx = nodes[i].x - nodes[j].x;
                const dy = nodes[i].y - nodes[j].y;
                const dist = Math.sqrt(dx * dx + dy * dy);

                if (dist < 150) {
                    ctx.beginPath();
                    ctx.moveTo(nodes[i].x, nodes[i].y);
                    ctx.lineTo(nodes[j].x, nodes[j].y);
                    
                    const opacity = (1 - dist / 150) * 0.1;
                    const color = props.processing ? '255, 0, 234' : '0, 242, 255';
                    ctx.strokeStyle = `rgba(${color}, ${opacity})`;
                    ctx.stroke();
                }
            }
        }

        nodes.forEach(node => {
            node.update();
            node.draw(ctx, props.processing);
        });

        animationId = requestAnimationFrame(animate);
    };

    animate();
};

onMounted(() => {
    initCanvas();
    window.addEventListener('resize', initCanvas);
    canvasRef.value?.addEventListener('mousemove', handleMouseMove);
    canvasRef.value?.addEventListener('mouseleave', handleMouseLeave);
});

onUnmounted(() => {
    cancelAnimationFrame(animationId);
    window.removeEventListener('resize', initCanvas);
    canvasRef.value?.removeEventListener('mousemove', handleMouseMove);
    canvasRef.value?.removeEventListener('mouseleave', handleMouseLeave);
});
</script>

<template>
    <canvas ref="canvasRef" class="w-full h-full opacity-40"></canvas>
</template>

<style scoped>
canvas {
    filter: blur(0.5px);
}
</style>
